package log

import (
	"errors"
)

// ErrLoggerAlreadySetup is the error, that will be returned,
// in the case that the logger has already been setup once
// before.
var ErrLoggerAlreadySetup = errors.New("the logger has already been setup")

// iLogger is a singleton to the actual logging instance
var iLogger internalLogger = newInternalLogger()

// builder handles all the configurations for
// the global logger.
type builder struct {
	handlers []LogHandler
}

// NewBuilder will create a builder that is used to create
// and manage the global logger instance.
func NewBuilder() builder {
	return builder{
		handlers: make([]LogHandler, handlerStartLen, handlerStartCap),
	}
}

// RegisterHandler will register a logging handler to the builder.
// This will not yet register the handler with the global logger.
func (b *builder) RegisterHandler(handler BuildLogHandler) error {
	lh, err := handler.Build()
	if err != nil {
		return err
	}
	// filter out the NONE logger levels as they are
	// not really valid.
	if !lh.checkLevel(None) {
		b.handlers = append(b.handlers, lh)
	}
	return nil
}

// Build will setup the logger, but only a single time,
// if the logger has already been setup before, then
// this will return an error.
func (b *builder) Build() error {
	return iLogger.copy(b.handlers)
}

// Reset will forcefully reset the configuration
// of the global logger.
// Attention this does not reset the builder
// configuration.
// Attention use this function with caution as
// might create unexpected issues if used. It
// should only be used for tests.
func (b *builder) Reset() {
	iLogger.reset()
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
	log.Log(Warn, v...)
}

// Infof will write and format an info message.
func (log logger) Warnf(format string, v ...interface{}) {
	log.Logf(Warn, format, v...)
}

// Error will write an error message.
func (log logger) Error(v ...interface{}) {
	log.Log(Error, v...)
}

// Errorf will write and format a error level message.
func (log logger) Errorf(format string, v ...interface{}) {
	log.Logf(Error, format, v...)
}
