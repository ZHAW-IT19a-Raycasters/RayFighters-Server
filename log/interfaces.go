package log

// LogHandler is the actual logger
// it handles all the from the
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
