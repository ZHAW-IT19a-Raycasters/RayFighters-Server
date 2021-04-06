package log

const (
	handlerStartCap = 3
	handlerStartLen = 0
)

type internalLogger struct {
	isSetup  bool
	handlers []LogHandler
}

func newInternalLogger() internalLogger {
	return internalLogger{
		isSetup: false,
	}
}

func (lg *internalLogger) copy(from []LogHandler) error {
	if lg.isSetup {
		return ErrLoggerAlreadySetup
	}

	// copying the slice, so that there is no change
	// of changing anything after assigment
	lg.handlers = make([]LogHandler, len(from))
	copy(lg.handlers, from)

	lg.isSetup = true
	return nil
}

func (lg *internalLogger) log(level Level, v ...interface{}) {
	for _, handler := range lg.handlers {
		handler.log(level, v...)
	}
}

func (lg internalLogger) logf(level Level, format string, v ...interface{}) {
	for _, handler := range lg.handlers {
		handler.logf(level, format, v...)
	}
}
