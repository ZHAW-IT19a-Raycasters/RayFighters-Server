package log

import (
	"errors"
)

// logFile is the file, where to the logs shall be logged to
const logFile string = "logs.txt"

// ErrLoggerAlreadySetup is the error, that will be returned,
// in the case that the logger has already been setup once
// before.
var ErrLoggerAlreadySetup = errors.New("the logger has already been setup")

// iLogger is a singleton to the actual logging instance
var iLogger internalLogger = newInternalLogger()

type builder struct {
	handlers []LoggingHandler
}

func NewBuilder() builder {
	return builder{
		handlers: make([]LoggingHandler, handlerStartLen, handlerStartCap),
	}
}

func (b *builder) RegisterHandler(handler LoggingHandler) {
	b.handlers = append(b.handlers, handler)
}

// Build will setup the logger, but only a single time,
// if the logger has already been setup before, then
// this will return an error.
func (b *builder) Build() error {
	return iLogger.copy(b.handlers)
}

// logger is a pseudo wrapper around the internal logger singleton.
// It does not contain any values, but the methods written for the
// logger all point to the internal logger variable one.
type logger struct{}

// Log returns a copy of the logger.
func Log() logger {
	return logger{}
}

// Log logs all the levels requested.
func (log logger) Log(level Level, v ...interface{}) {
	iLogger.log(level, v...)
}

// Logf logs all the levels and formats the interfaces values.
func (log logger) Logf(level Level, format string, v ...interface{}) {
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
