package log

import "sync"

const (
	handlerStartCap = 3
	handlerStartLen = 0
)

type internalLogger struct {
	lock     sync.Mutex
	isSetup  bool
	handlers []LogHandler
}

func newInternalLogger() internalLogger {
	return internalLogger{
		handlers: nil,
		lock:     sync.Mutex{},
		isSetup:  false,
	}
}

func (lg *internalLogger) copy(from []LogHandler) error {
	if lg.isSetup {
		return ErrLoggerAlreadySetup
	}

	lg.lock.Lock()
	defer lg.lock.Unlock()

	// copying the slice, so that there is no change
	// of changing anything after assigment
	lg.handlers = make([]LogHandler, len(from))
	copy(lg.handlers, from)

	lg.isSetup = true

	return nil
}

func (lg *internalLogger) reset() {
	iLogger.lock.Lock()
	defer iLogger.lock.Unlock()
	lg.handlers = nil
	lg.isSetup = false
}

func (lg *internalLogger) log(level Level, v ...interface{}) {
	for _, handler := range lg.handlers {
		handler.log(level, v...)
	}
}

func (lg *internalLogger) logf(level Level, format string, v ...interface{}) {
	for _, handler := range lg.handlers {
		handler.logf(level, format, v...)
	}
}
