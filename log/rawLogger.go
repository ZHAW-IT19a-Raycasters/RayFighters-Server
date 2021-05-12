package log

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

type rawLogger struct {
	ilogger *log.Logger
	level   Level
}

// newRawLogger will create the actual internal logger instance,
// used for all logging.
func newRawLogger(level Level, out io.Writer) *rawLogger {
	// using 0 as flag to get full controll
	lg := log.New(out, "", 0)
	lg.SetOutput(out)
	return &rawLogger{
		ilogger: lg,
		level:   level,
	}
}

// Level will return the level used by the raw logger
func (log rawLogger) Level() Level {
	return log.level
}

// checkLevel is the function that checks if the requested level
// is allowed to be used per configuration.
func (log rawLogger) checkLevel(level Level) bool {
	return level >= log.Level()
}

// Log logs all the levels requested.
func (log rawLogger) log(level Level, v ...interface{}) {

	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	// insert before the values
	v = prependLevel(level, v)

	log.ilogger.Println(v...)
}

// Logf logs all the levels and formats the interfaces values.
func (log rawLogger) logf(level Level, format string, v ...interface{}) {
	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	// using a string builder for faster concatination
	var builder strings.Builder

	// calculate the base size of the builder
	// to avoid some unneded allocations
	size := 0

	size += 2 + 30           // formated date time
	size += 1                // spacing between date and level
	size += 2 + maxPrefixLen // prefix like [DEBUG] plus size padding
	size += 1                // spacing between base and prefix

	builder.Grow(size)

	// prepare the date time now
	now := time.Now()
	zone, _ := now.Zone()
	fmt.Fprintf(&builder, "[%d-%02d-%02d %02d:%02d:%02d %04d  % 4s]", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), zone)

	fmt.Fprintf(&builder, " [%s]", level.String())
	neededLen := maxPrefixLen - len(level.String())

	// fill up the string with the missing spacing
	for i := 0; i < neededLen; i += 1 {
		fmt.Fprint(&builder, " ")
	}

	// add the user defined format
	fmt.Fprintf(&builder, format, v...)

	//log.ilogger.Printf(fmt.Sprintf("[%s] %s %s", date, level.String(), format), v...)

	log.ilogger.Println(builder.String())
}
