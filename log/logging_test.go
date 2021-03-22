package log

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func setup(level Level, t *testing.T) bytes.Buffer {
	rawBuf := make([]byte, 0, 500)

	buffer := bytes.NewBuffer(rawBuf)
	iLogger = newInternalLogger(level, buffer)
	t.Cleanup(func() {
		iLogger = nil
	})
	return *buffer
}
func TestLogger(t *testing.T) {
	buffer := setup(Debug, t)

	logger := Log()

	logger.Debug("help me")

	const expetedResult string = "DEBUG help me"
	res := buffer.String()

	if strings.HasSuffix(res, expetedResult) {
		t.Errorf("The buffer should be filled with %s", expetedResult)
	}

}
func TestLoggerNoDebug(t *testing.T) {
	buffer := setup(Info, t)

	buffer.WriteRune(' ')

	logger := Log()
	logger.Debug("help me")

	const expetedResult string = "DEBUG help me"
	res := buffer.String()

	if strings.HasSuffix(res, expetedResult) {
		t.Errorf("The buffer should be not be filled with %s", expetedResult)
	}
}

func TestDoubleLoggingInit(t *testing.T) {
	t.Cleanup(func() {
		// remove the newly created log file, as it is only used for
		// testing anyway.
		os.Remove(logFile)
	})
	defer func() {
		r := recover()
		if r == nil {
			t.Error("The code should panic on a double initialization.")
		}
	}()

	Setup(Debug)
	Setup(Debug)
}
