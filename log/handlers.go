package log

import (
	"errors"
	"io"
	"os"
)

// ErrLevelIsNone is returned, when the builder can not create the
// object due to the level constraints.
var ErrLevelIsNone = errors.New("handler level is none")

// checkNone will check if the given level is none or not.
func checkNone(level Level) bool {
	return level == None
}

// WriterHandler is a handler used as a base for
// for any io.Writer loggers.
type WriterHandler struct {
	// level is the level limit
	level Level
	// out is the exit point to which
	// the logs are propagated.
	out io.Writer
}

// NewWriterHandler will create a new writer handler
// based on the writer parameter.
func NewWriterHandler(level Level, writer io.Writer) WriterHandler {
	return WriterHandler{
		level: level,
		out:   writer,
	}
}

// Build will create the internal logger handler
// that can then be added to the system.
func (wh WriterHandler) Build() (LogHandler, error) {
	if checkNone(wh.level) {
		return nil, ErrLevelIsNone
	}
	// check for none
	return newRawLogger(wh.level, wh.out), nil

}

// ConsoleHandler will log to the console.
type ConsoleHandler struct {
	WriterHandler
}

// NewConsoleHandler will create a new console handler
// and set it up to log the given level on.
func NewConsoleHandler(level Level) ConsoleHandler {
	return ConsoleHandler{
		WriterHandler: NewWriterHandler(level, os.Stdout),
	}
}

// FileHandler is the handler struct for logging to
// the file system.
type FileHandler struct {
	// openFile is the os functionality used for opening files.
	// The function is stored here so that it is possible to mock
	// it and get 100% coverage.
	openFile func() (*os.File, error)
	// fileName where to the log will be stored to
	fileName string
	// level is the level limit
	level Level
}

// NewFileHandler will create and open if needed a
// logging file on the file system.
func NewFileHandler(level Level, fileName string) FileHandler {
	return FileHandler{
		fileName: fileName,
		level:    level,
		openFile: func() (*os.File, error) {
			return os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		},
	}
}

// Level will return the minimal level the handler
// log from.
func (fh FileHandler) Level() Level {
	return fh.level
}

// FileName will return the file location to which
// the log handler will propagate the logs to.
func (fh FileHandler) FileName() string {
	return fh.fileName
}

// Build will create the actual logging instance.
// And if needed create the file on the file system
// as well.
func (fh FileHandler) Build() (LogHandler, error) {
	if checkNone(fh.level) {
		return nil, ErrLevelIsNone
	}

	out, err := fh.openFile()
	if err != nil {
		return nil, err
	}
	return newRawLogger(fh.Level(), out), nil
}
