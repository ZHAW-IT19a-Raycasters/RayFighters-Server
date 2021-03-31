package log

import (
	"io"
	"os"
)

type WriterHandler struct {
	level Level
	out   io.Writer
}

func NewWriterHandler(level Level, writer io.Writer) WriterHandler {
	return WriterHandler{
		level: level,
		out:   writer,
	}
}

func (wh WriterHandler) Build() (LogHandler, error) {
	return newRawLogger(wh.level, wh.out), nil

}

type ConsoleHandler struct {
	WriterHandler
}

func NewConsoleHandler(level Level) ConsoleHandler {
	return ConsoleHandler{
		WriterHandler: NewWriterHandler(level, os.Stdout),
	}
}

type FileHandler struct {
	// openFile is the os functionality used for opening files.
	// The function is stored here so that it is possible to mock
	// it and get 100% coverage.
	openFile func() (*os.File, error)
	// fileName where to the log will be stored to
	fileName string
	level    Level
}

func NewFileHandler(level Level, fileName string) FileHandler {
	return FileHandler{
		fileName: fileName,
		level:    level,
		openFile: func() (*os.File, error) {
			return os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		},
	}
}

func (fh FileHandler) Level() Level {
	return fh.level
}

func (fh FileHandler) FileName() string {
	return fh.fileName
}

func (fh FileHandler) Build() (LogHandler, error) {
	out, err := fh.openFile()
	if err != nil {
		return nil, err
	}
	return newRawLogger(fh.Level(), out), nil
}
