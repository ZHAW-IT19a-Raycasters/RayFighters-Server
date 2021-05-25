package log

import "sync"

const (
	handlerStartCap = 3
	handlerStartLen = 0
)

// internalLogger is the internal instance
// of the logger that is responsible for
// multiplexing the given debug information.
type internalLogger struct {
	lock     sync.RWMutex
	isSetup  bool
	handlers []LogHandler
}

// newInternalLogger will as the name implies
// create a new internal logger object.
func newInternalLogger() internalLogger {
	return internalLogger{
		handlers: nil,
		lock:     sync.RWMutex{},
		isSetup:  false,
	}
}

// copy copy the log handler from the parameter
// into the logger setup.
func (lg *internalLogger) copy(from []LogHandler) error {
	lg.lock.Lock()
	defer lg.lock.Unlock()
	if lg.isSetup {
		return ErrLoggerAlreadySetup
	}

	// copying the slice, so that there is no change
	// of changing anything after assignment
	lg.handlers = make([]LogHandler, len(from))
	copy(lg.handlers, from)

	lg.isSetup = true

	return nil
}

// reset will force an internal reset of all values.
func (lg *internalLogger) reset() {
	iLogger.lock.Lock()
	defer iLogger.lock.Unlock()
	lg.handlers = nil
	lg.isSetup = false
}

// log will log a message of the given type.
func (lg *internalLogger) log(level Level, v ...interface{}) {
	lg.lock.RLock()
	defer lg.lock.RUnlock()
	// this is valid as if the handler array is nil
	// it is translated to be of length zero, so
	// this loop will do nothing
	for _, handler := range lg.handlers {
		handler.log(level, v...)
	}
}

// logf will first format and then log a given message.
func (lg *internalLogger) logf(level Level, format string, v ...interface{}) {
	lg.lock.RLock()
	defer lg.lock.RUnlock()
	// this is valid as if the handler array is nil
	// it is translated to be of length zero, so
	// this loop will do nothing
	for _, handler := range lg.handlers {
		handler.logf(level, format, v...)
	}
}
