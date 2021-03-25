package log

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func cleanUpTmpFile(name string) {
	iLogger = nil
	if _, err := os.Stat(name); errors.Is(err, os.ErrExist) {
		os.Remove(name)
	}
}

func setupTmpFile(level Level, t *testing.T) (string, error) {
	tmpFile, err := ioutil.TempFile("", "log*.txt")
	if err != nil {
		t.Errorf("on creating the tmp file, this happend [%s]", err.Error())
		return "", err
	}

	err = SetupLogFile(level, tmpFile.Name())
	if err != nil {
		t.Errorf("on creating the logger, this happend [%s]", err.Error())
		return "", err
	}

	t.Cleanup(func() {
		cleanUpTmpFile(tmpFile.Name())
	})
	return tmpFile.Name(), nil
}

func setupBuffer() bytes.Buffer {
	rawBuf := make([]byte, 0, 500)

	buffer := bytes.NewBuffer(rawBuf)
	return *buffer
}

func setupLogger(level Level, out io.Writer, t *testing.T) {
	setILogger(newInternalLogger(level, out))
	t.Cleanup(func() {
		resetILogger()
	})
}

func setupBufferLogger(level Level, t *testing.T) bytes.Buffer {
	buffer := setupBuffer()
	setupLogger(level, &buffer, t)

	return buffer
}

func TestLogger(t *testing.T) {
	buff := setupBuffer()
	buff.WriteByte(' ')

	lg := Log()
	lg.Debug("Something")
	lg.Debugf("Something %s", "2")
	// nothing happens here
	res := buff.String()
	if res != " " {
		t.Error("Something unexpected happend with the buffer")
		return
	}
	// nothing happens here
	test := func(check string) bool {
		res = buff.String()
		res = strings.TrimSpace(res)
		if !strings.HasSuffix(res, check) {
			t.Error("Wrong values have been send by the logger")
			return false
		}
		return true
	}
	// remove values from buff
	buff.Reset()

	setupLogger(Debug, &buff, t)

	lg.Debug("Something")
	if !test("Something") {
		return
	}

	lg.Debugf("Something [%s]", "2")
	if !test("Something [2]") {
		return
	}
}

func TestLoggerDebug(t *testing.T) {

	buffer := setupBufferLogger(Debug, t)

	logger := Log()

	logger.Debug("help me")

	const expetedResult string = "DEBUG help me"
	res := buffer.String()

	if strings.HasSuffix(res, expetedResult) {
		t.Errorf("The buffer should be filled with %s", expetedResult)
	}
}

func TestLoggerNoDebug(t *testing.T) {
	buffer := setupBufferLogger(Info, t)

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
	// Attention this calles Setup already
	_, err := setupTmpFile(Debug, t)
	if err != nil {
		t.Errorf("There should have not been an error here.\n%s", err.Error())
		return
	}
	err = Setup(Debug)
	if err == nil {
		t.Error("The error for a double setup is missing.")
	}
}

func TestLoggingToFile(t *testing.T) {
	filename, err := setupTmpFile(Debug, t)
	if err != nil {
		return
	}

	Setup(Debug)
	log := Log()
	log.Infof("the value %d", 42)

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Unexpected file error happend.\n%s", err.Error())
		return
	}
	str := strings.TrimSpace(string(data))
	if !strings.HasSuffix(str, "the value 42") {
		t.Errorf("an incorrect suffix was used, for line\n[%s]", str)
	}
}
