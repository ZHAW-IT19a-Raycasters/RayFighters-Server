package log

import (
	"fmt"
	"io"
	"log"
)

type rawLogger struct {
	ilogger *log.Logger
	level   Level
}

// newRawLogger will create the actual internal logger instance,
// used for all logging.
func newRawLogger(level Level, out io.Writer) *rawLogger {
	return &rawLogger{
		ilogger: log.New(out, "", log.Ldate|log.Ltime|log.Lshortfile),
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

	log.ilogger.Printf(fmt.Sprintf("%s %s", level.String(), format), v...)
}
