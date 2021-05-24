package log

import (
	"fmt"
	"strings"
)

// maxPrefixLen is the maximal length of the level names
// this is used for calculating the required size
// for the level strings.
const maxPrefixLen int = 5

// Level is an enum representation of the
// currently supported levels
type Level int

const (
	// None is the lowest level supported although
	// it deactivates logging rather then enables
	// it.
	None Level = iota
	// Debug is the lowest possible level of messages
	// to be logged, it will add a "DEBUG" prefix
	// to the given message.
	Debug
	// Info is the second log level.
	// It will add a "INFO" prefix.
	Info
	// Warn is the third log level.
	// It will add the "WARN" prefix.
	Warn
	// Error is the highest logging level.
	// It will add the "ERROR" prefix.
	Error
)

const (
	noneStr  string = "NONE"
	debugStr string = "DEBUG"
	infoStr  string = "INFO"
	warnStr  string = "WARN"
	errorStr string = "ERROR"
)

// overview contains an iterable list of all
// the constant values.
var overview = [...]string{
	noneStr, debugStr, infoStr, warnStr, errorStr,
}

// ErrUnknownLevel is the error returned when
// we were unable to detect the level string.
type ErrUnknownLevel struct {
	given string
}

// Error represents the error interface
func (err ErrUnknownLevel) Error() string {
	return fmt.Sprintf("unknown level name [%s] used, there are only these supported [%s]", err.given, strings.Join(overview[:], ", "))
}

// Is satisfies the errors.Is definition
func (err ErrUnknownLevel) Is(target error) bool {
	_, ok := target.(ErrUnknownLevel)
	return ok
}

// String will return the string representation
func (l Level) String() string {
	// this is safe as it will crash if
	// an illegal level is used.
	return overview[l]
}

// LevelDetect will convert the string representation
// into its Level. This will return an error if there
// was ain incorrect string used.
func LevelDetect(level string) (Level, error) {
	processedStr := strings.TrimSpace(level)
	processedStr = strings.ToUpper(processedStr)

	for i, lvStr := range overview {
		if lvStr == processedStr {
			return Level(i), nil
		}
	}

	return Error, ErrUnknownLevel{
		given: level,
	}
}
