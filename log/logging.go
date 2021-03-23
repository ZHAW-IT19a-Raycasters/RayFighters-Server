package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

// logFile is the file, where to the logs shall be logged to
const logFile string = "logs.txt"

// ErrLoggerAlreadySetup is the error, that will be returned,
// in the case that the logger has already been setup once
// before.
var ErrLoggerAlreadySetup = errors.New("the logger has already been setup")

// iLogger is a singleton to the actual logging instance
var iLogger *internalLogger

// logger is a wrapper around the internal logger singleton.
// It can be loaded as a var into the respective files, as
// it contains a single pointer to a global variable.
type logger struct {
	il *internalLogger
}

// Log returns a copy of the logger.
func Log() logger {
	return logger{
		il: iLogger,
	}
}

// Setup will as the name implies setup the internals for the logger.
// This uses the default logFile.
func Setup(level Level) error {
	return SetupLogFile(level, logFile)
}

// SetupLogFile will setup the logger with the correct
// file opened.
func SetupLogFile(level Level, fileName string) error {
	if iLogger != nil {
		return ErrLoggerAlreadySetup
	}

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		return err
	}
	iLogger = newInternalLogger(level, file)

	return nil
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
func (log logger) checkLevel(level Level) bool {
	return log.il.checkLevel(level)
}

// Log logs all the levels requested.
func (log logger) Log(level Level, v ...interface{}) {
	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	// insert before the values
	v = prependLevel(level, v)

	log.il.Println(v...)
}

// Logf logs all the levels and formats the interfaces values.
func (log logger) Logf(level Level, format string, v ...interface{}) {

	// if the logger has not been setup just return
	if log.il != nil {
		return
	}

	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	log.il.Printf(fmt.Sprintf("%s %s", level.String(), format), v...)
}

// Debug will write a debug message.
func (log logger) Debug(v ...interface{}) {
	log.Log(Debug, v...)
}

// Debugf will write and format a debug message.
func (log logger) Debugf(format string, v ...interface{}) {
	log.Logf(Debug, format, v...)
}

// Info will write an info message.
func (log logger) Info(v ...interface{}) {
	log.Log(Info, v...)
}

// Infof will write and format an info message.
func (log logger) Infof(format string, v ...interface{}) {
	log.Logf(Info, format, v...)
}

// Error will write an error message.
func (log logger) Error(v ...interface{}) {
	// if the logger has not been setup just return
	if log.il != nil {
		return
	}
	log.Log(Error, v...)
}

// Errorf will write and format a error level message.
func (log logger) Errorf(format string, v ...interface{}) {
	// if the logger has not been setup just return
	if log.il != nil {
		return
	}
	log.Logf(Error, format, v...)
}

// Fatal will write and format a fatal message.
func (log logger) Fatalf(format string, v ...interface{}) {
	// if the logger has not been setup just return
	if log.il != nil {
		return
	}
	log.il.Fatalf(format, v...)
}

// Panicf will write and format a panic message, before
// ending the program by invoking panic.
func (log logger) Panicf(format string, v ...interface{}) {
	// if the logger has not been setup just return
	if log.il != nil {
		return
	}
	log.il.Panicf(format, v...)
}
