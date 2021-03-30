package log

import (
	"fmt"
	"io"
    "os"
	"log"
)

type LoggingHandler interface {
    checkLevel(level Level) bool    
    log(level Level, v...interface{})
    logf(level Level, format string, v ...interface{})
    panicf(format string, v...interface{})
}


type internalLoggerWrapper struct {
    handlers []LoggingHandler
}

func (ilw *internalLoggerWrapper) RegisterHandler(handler LoggingHandler) {
    ilw.handlers = append(ilw.handlers, handler)
}

type ConsoleHandler struct {
    *internalLogger
}

func NewConsoleHandler(level Level) (ConsoleHandler, error){
    // TODO:
    return ConsoleHandler{}, nil
}


type FileHandler struct {
    fileName string
    *internalLogger
}


func NewFileHandler(fileName string, level Level) (FileHandler, error) {
	file, err := openFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		return FileHandler{}, err
	}   

    return FileHandler {
        fileName: fileName,
        internalLogger: newInternalLogger(level, file),
    }, nil
}



type internalLogger struct {
	ilogger *log.Logger
	level   Level
}

// newInternalLogger will create the actual internal logger instance,
// used for all logging.
func newInternalLogger(level Level, out io.Writer) *internalLogger {
	return &internalLogger{
		ilogger: log.New(out, "", log.Ldate|log.Ltime|log.Lshortfile),
		level:   level,
	}
}

// checkLevel is the function that checks if the requested level
// is allowed to be used per configuration.
func (log internalLogger) checkLevel(level Level) bool {
	return level >= log.level
}

// Log logs all the levels requested.
func (log internalLogger) log(level Level, v ...interface{}) {

	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	// insert before the values
	v = prependLevel(level, v)

	log.ilogger.Println(v...)
}

// Logf logs all the levels and formats the interfaces values.
func (log internalLogger) logf(level Level, format string, v ...interface{}) {
	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	log.ilogger.Printf(fmt.Sprintf("%s %s", level.String(), format), v...)
}

// panicf will log the line then panic the system.
func (log internalLogger) panicf(format string, v ...interface{}) {
    // TODO: figure out how to propagate this panic
    // possibly repepanic per panic?
	log.ilogger.Panicf(format, v...)
}

