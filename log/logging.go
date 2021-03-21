package log

import (
	"io"
	"log"
	"os"
)

const logFile string = "logs.txt"

type Level int

const (
	Debug Level = iota
	Info
	Warning
	Error
)

const (
	debugStr string = "DEBUG"
	infoStr  string = "INFO"
	warnStr  string = "WARNING"
	errorStr string = "ERROR"
)

func (l Level) String() string {
	switch l {
	case Debug:
		return debugStr
	case Info:
		return infoStr
	case Warning:
		return warnStr
	case Error:
		return errorStr
	default:
		panic("Unsupported level used")
	}
}

type logger struct {
	ilogger *log.Logger
	level   Level
}

var loggerV *logger

func Log() (logger *logger) {
	return loggerV
}

func Setup(level Level) error {
	if loggerV != nil {
		panic("The logger may only be setup once!")
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		return err
	}

	loggerV = newLogger(level, file)

	return nil
}

func newLogger(level Level, out io.Writer) *logger {
	return &logger{
		ilogger: log.New(out, "", log.Ldate|log.Ltime|log.Lshortfile),
		level:   level,
	}
}

func (log *logger) checkLevel(level Level) bool {
	return level >= log.level
}

func prependLevel(y Level, x []interface{}) []interface{} {
	x = append(x, 0)
	copy(x[1:], x)
	x[0] = y.String()
	return x
}

func (log *logger) Log(level Level, v ...interface{}) {
	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	// insert before the values
	v = prependLevel(level, v)

	log.ilogger.Println(v...)
}

func (log *logger) Logf(level Level, format string, v ...interface{}) {
	// check if logged level is lower of higher then the set version
	if !log.checkLevel(level) {
		return
	}

	// insert before the values
	v = prependLevel(level, v)

	log.ilogger.Printf(format, v...)
}

func (log *logger) Debug(v ...interface{}) {
	log.Log(Debug, v...)
}

func (log *logger) Debugf(format string, v ...interface{}) {
	log.Logf(Debug, format, v...)
}

func (log *logger) Info(v ...interface{}) {
	log.Log(Info, v...)
}

func (log *logger) Error(v ...interface{}) {
	log.Log(Error, v...)
}

func (log *logger) Errorf(format string, v ...interface{}) {
	log.Logf(Error, format, v...)
}

func (log *logger) Fatal(v ...interface{}) {
	log.ilogger.Fatal(v...)
}

func (log *logger) Fatalf(format string, v ...interface{}) {
	log.ilogger.Fatalf(format, v...)
}

func (log *logger) Panic(v ...interface{}) {
	log.ilogger.Panic(v...)
}

func (log *logger) Panicf(format string, v ...interface{}) {
	log.ilogger.Panicf(format, v...)
}
