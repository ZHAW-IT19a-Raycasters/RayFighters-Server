package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// rawLogger is the most inner data structure
// that does directly interact with the
// underlying logger structure.
type rawLogger struct {
	ilogger *log.Logger
	level   Level
	// bPool will help out with not having to handle as many
	// byte buffer allocations during a busy period
	bPool sync.Pool
}

// newRawLogger will create the actual internal logger instance,
// used for all logging.
func newRawLogger(level Level, out io.Writer) *rawLogger {
	// using 0 as flag to get no prepared output
	return &rawLogger{
		ilogger: log.New(out, "", 0),
		level:   level,
		bPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// Level will return the level used by the raw logger
func (log *rawLogger) Level() Level {
	return log.level
}

// checkLevel is the function that checks if the requested level
// is allowed to be used per configuration.
func (log *rawLogger) checkLevel(level Level) bool {
	// using the fact that the level enum represents an integer
	// to simplify the check functionality.
	return level >= log.Level()
}

// prepareBuilder will prepare a formatter string used for logging
// using a custom format.
func (log *rawLogger) prepareBuilder(level Level) *bytes.Buffer {
	// using a string builder for faster concatenation
	builder := log.bPool.Get().(*bytes.Buffer)
	builder.Reset()
	// calculate the base size of the builder
	// to avoid some unneeded allocations
	size := 0

	size += 2 + 30           // formatted date time
	size += 1                // spacing between date and level
	size += 2 + maxPrefixLen // prefix like [DEBUG] plus size padding
	size += 1                // spacing between base and prefix

	builder.Grow(size)

	// prepare the date time now
	now := time.Now()
	zone, _ := now.Zone()
	fmt.Fprintf(builder, "[%d-%02d-%02d %02d:%02d:%02d %04d  % 4s]", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), zone)

	fmt.Fprintf(builder, " [%s]", level.String())
	// + 1 is needed for the spacing at the end
	neededLen := maxPrefixLen - len(level.String()) + 1

	// fill up the string with the missing spacing
	for i := 0; i < neededLen; i += 1 {
		fmt.Fprint(builder, " ")
	}

	return builder
}

// print will write the builder out to the internal logging
// instance.
func (log *rawLogger) print(builder *bytes.Buffer) {
	log.ilogger.Println(builder.String())
	log.bPool.Put(builder)
}

// Log logs all the levels requested.
func (log *rawLogger) log(level Level, v ...interface{}) {

	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	builder := log.prepareBuilder(level)

	fmt.Fprint(builder, v...)

	log.print(builder)
}

// Logf logs all the levels and formats the interfaces values.
func (log *rawLogger) logf(level Level, format string, v ...interface{}) {
	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	builder := log.prepareBuilder(level)

	// add the user defined format
	fmt.Fprintf(builder, format, v...)

	log.print(builder)
}
