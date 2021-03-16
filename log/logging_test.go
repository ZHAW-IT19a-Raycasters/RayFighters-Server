package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrependlevel(t *testing.T) {
	org := []interface{}{"1", "2", "3"}
	expected := []interface{}{Debug.String(), "1", "2", "3"}
	res := prependLevel(Debug, org)
	if len(res) != len(expected) {
		t.Error("expected lenght not reached")
	}
	for i := range res {
		if res[i] != expected[i] {
			t.Errorf("[%s] is not equals to [%s]", res[i], expected[i])
		}
	}
}

func TestLogger(t *testing.T) {
	rawBuf := make([]byte, 0, 500)

	buffer := bytes.NewBuffer(rawBuf)

	logger := newLogger(Debug, buffer)
	logger.Debug("help me")

	const expetedResult string = "DEBUG help me"
	res := buffer.String()

	if strings.HasSuffix(res, expetedResult) {
		t.Errorf("The buffer should be filled with %s", expetedResult)
	}

}

func TestLoggerNoDebug(t *testing.T) {
	rawBuf := make([]byte, 0, 500)

	buffer := bytes.NewBuffer(rawBuf)
	buffer.WriteRune(' ')

	logger := newLogger(Info, buffer)
	logger.Debug("help me")

	const expetedResult string = "DEBUG help me"
	res := buffer.String()

	if strings.HasSuffix(res, expetedResult) {
		t.Errorf("The buffer should be not be filled with %s", expetedResult)
	}
}

func TestDoubleLoggingInit(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("The code should panic on a double initialization.")
		}
	}()

	Setup(Debug)
	Setup(Debug)
}
