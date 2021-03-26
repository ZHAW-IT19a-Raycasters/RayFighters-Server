package log

import (
	"fmt"
	"io"
	"log"
)

type internalLogger struct {
	ilogger *log.Logger
	level   Level
}

// newInternalLogger will create the actual internal logger instance,
// used for all logging.
func newInternalLogger(level Level, out io.Writer) *internalLogger {
	return &internalLogger{
		ilogger: log.New(out, "", log.Ldate|log.Ltime|log.Lshortfile),
		level:   level,
	}
}

// checkLevel is the function that checks if the requested level
// is allowed to be used per configuration.
func (log internalLogger) checkLevel(level Level) bool {
	return level >= log.level
}

// Log logs all the levels requested.
func (log internalLogger) log(level Level, v ...interface{}) {

	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	// insert before the values
	v = prependLevel(level, v)

	log.ilogger.Println(v...)
}

// Logf logs all the levels and formats the interfaces values.
func (log internalLogger) logf(level Level, format string, v ...interface{}) {
	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	log.ilogger.Printf(fmt.Sprintf("%s %s", level.String(), format), v...)
}

// panicf will log the line then panic the system.
func (log internalLogger) panicf(format string, v ...interface{}) {
	log.ilogger.Panicf(format, v...)
}
