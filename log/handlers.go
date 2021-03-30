package log

import (
	"io"
	"io/fs"
	"os"
)

type WriterHandler struct {
	*rawLogger
}

func NewWriterHandler(level Level, writer io.Writer) WriterHandler {
	return WriterHandler{
		rawLogger: newRawLogger(level, writer),
	}
}

type ConsoleHandler struct {
	*rawLogger
}

func NewConsoleHandler(level Level) ConsoleHandler {
	return ConsoleHandler{
		rawLogger: newRawLogger(level, os.Stdout),
	}
}

type FileHandler struct {
	// openFile is the os functionality used for opening files.
	// The function is stored here so that it is possible to mock
	// it and get 100% coverage.
	openFile func(name string, flag int, perm fs.FileMode) (*os.File, error)
	// fileName where to the log will be stored to
	fileName string
	*rawLogger
}

func NewFileHandler(fileName string, level Level) (FileHandler, error) {
	fh := FileHandler{
		fileName: fileName,
	}

	file, err := fh.openFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		return fh, err
	}

	fh.rawLogger = newRawLogger(level, file)

	return fh, nil
}

func (fh *FileHandler) FileName() string {
	return fh.fileName
}
