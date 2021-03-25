package log

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

// String will return the strign representation
func (l Level) String() string {
	msg := [...]string{
		debugStr, infoStr, warnStr, errorStr,
	}
	return msg[l]
}

// prependLevel is a helper that inserts the level name before
// the other values.
func prependLevel(y Level, x []interface{}) []interface{} {
	// force realloc
	x = append(x, 0)
	// move all the values in x to the end
	copy(x[1:], x)
	// add the important information to the front
	x[0] = y.String()
	return x
}
