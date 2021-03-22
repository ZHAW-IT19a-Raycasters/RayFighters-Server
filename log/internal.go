package log

import "log"

type internalLogger struct {
	ilogger *log.Logger
	level   Level
}

func (log internalLogger) checkLevel(level Level) bool {
	return level >= log.level
}

func (log internalLogger) Println(v ...interface{}) {
	log.ilogger.Println(v...)
}

func (log internalLogger) Printf(format string, v ...interface{}) {
	log.ilogger.Printf(format, v...)
}

func (log internalLogger) Fatalf(format string, v ...interface{}) {
	log.ilogger.Fatalf(format, v...)
}

func (log internalLogger) Panicf(format string, v ...interface{}) {
	log.ilogger.Panicf(format, v...)
}
