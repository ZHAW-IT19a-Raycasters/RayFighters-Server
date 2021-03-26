package log

import (
	"errors"
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

// setILogger will set
func setILogger(lg *internalLogger) {
	iLogger = lg
}

// resetILogger will set the actall logger to nil.
func resetILogger() {
	iLogger = nil
}

// logger is a pseudo wrapper around the internal logger singleton.
// It does not contain any values, but the methods written for the
// logger all point to the internal logger variable one.
type logger struct{}

// Log returns a copy of the logger.
func Log() logger {
	return logger{}
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

	setILogger(newInternalLogger(level, file))

	return nil
}

func (log logger) isSetup() bool {
	return iLogger != nil
}

// Log logs all the levels requested.
func (log logger) Log(level Level, v ...interface{}) {
	// if the logger has not been setup just return
	if !log.isSetup() {
		return
	}
	iLogger.log(level, v...)
}

// Logf logs all the levels and formats the interfaces values.
func (log logger) Logf(level Level, format string, v ...interface{}) {
	// if the logger has not been setup just return
	if !log.isSetup() {
		return
	}
	iLogger.logf(level, format, v...)
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

// Info will write an info message.
func (log logger) Warn(v ...interface{}) {
	log.Log(Warning, v...)
}

// Infof will write and format an info message.
func (log logger) Warnf(format string, v ...interface{}) {
	log.Logf(Warning, format, v...)
}

// Error will write an error message.
func (log logger) Error(v ...interface{}) {
	log.Log(Error, v...)
}

// Errorf will write and format a error level message.
func (log logger) Errorf(format string, v ...interface{}) {
	log.Logf(Error, format, v...)
}

// Panicf will write and format a panic message, before
// ending the program by invoking panic.
func (log logger) Panicf(format string, v ...interface{}) {
	// if the logger has not been setup just return
	if !log.isSetup() {
		return
	}
	iLogger.panicf(format, v...)
}
