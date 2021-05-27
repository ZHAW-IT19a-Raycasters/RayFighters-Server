package log

// LogHandler is the actual logger interface
// it handles all the internal
// requirements for logging to the
// requested destination.
type LogHandler interface {
	checkLevel(level Level) bool
	log(level Level, v ...interface{})
	logf(level Level, format string, v ...interface{})
}

// LoggingHandler is the interface, that can create
// the loggers.
type BuildLogHandler interface {
	Build() (LogHandler, error)
}
