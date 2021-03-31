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

// resetILogger will renew the internal logger
// down to no configurations.
// Attention this function is used only for testing.
func resetILogger() {
	// remove all the previous configuration
	iLogger = newInternalLogger()
}

func cleanUpTmpFile(name string) {
	if _, err := os.Stat(name); errors.Is(err, os.ErrExist) {
		os.Remove(name)
	}
}

func setupTmpFile(level Level, t *testing.T) (*FileHandler, error) {
	tmpFile, err := ioutil.TempFile("", "log*.txt")
	if err != nil {
		t.Errorf("on creating the tmp file, this happend [%s]", err.Error())
		return nil, err
	}

	fh := NewFileHandler(level, tmpFile.Name())

	t.Cleanup(func() {
		cleanUpTmpFile(tmpFile.Name())
	})

	return &fh, nil
}

func setupBuffer() *bytes.Buffer {
	rawBuf := make([]byte, 0, 500)

	buffer := bytes.NewBuffer(rawBuf)
	return buffer
}

func setupWriteLogger(level Level, out io.Writer, t *testing.T) WriterHandler {
	return NewWriterHandler(level, out)
}

func setupBufferLogger(level Level, t *testing.T) (WriterHandler, *bytes.Buffer) {
	buffer := setupBuffer()
	wh := setupWriteLogger(level, buffer, t)

	return wh, buffer
}

func setupLogger(handlers []BuildLogHandler, t *testing.T) error {
	b := NewBuilder()

	for _, handler := range handlers {
		if err := b.RegisterHandler(handler); err != nil {
			return err
		}
	}

	err := b.Build()
	if err != nil {
		return err
	}

	t.Cleanup(resetILogger)

	return nil
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

	wh := setupWriteLogger(Debug, buff, t)
	err := setupLogger([]BuildLogHandler{wh}, t)
	if err != nil {
		t.Errorf("There was an error during logger setup\n%s", err.Error())

		return
	}

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

	wh, buffer := setupBufferLogger(Debug, t)

	err := setupLogger([]BuildLogHandler{wh}, t)
	if err != nil {
		t.Errorf("There was an error while setting up the logger\n%s", err.Error())
		return
	}

	logger := Log()

	logger.Debug("help me")

	const expetedResult string = "DEBUG help me"
	res := buffer.String()

	if strings.HasSuffix(res, expetedResult) {
		t.Errorf("The buffer should be filled with %s", expetedResult)
	}
}

func TestLoggerNoDebug(t *testing.T) {
	wh, buffer := setupBufferLogger(Info, t)

	err := setupLogger([]BuildLogHandler{wh}, t)
	if err != nil {
		t.Errorf("There was an error while setting up the logger\n%s", err.Error())
		return
	}

	check := func(f func() string) bool {
		buffer.Reset()
		exp := f()
		res := strings.TrimSpace(buffer.String())

		if strings.HasSuffix(res, exp) {
			t.Errorf("The buffer should be not be filled with %s", exp)
			return false
		}
		return true
	}

	logger := Log()
	debug := func() string {
		// doesn't do anything
		logger.Debug("help me")

		return "DEBUG help me"
	}

	debugf := func() string {

		logger.Debugf("help me%s", " - please")
		return "DEBUG help me - please"
	}

	for _, f := range [...]func() string{debug, debugf} {
		if !check(f) {
			return
		}
	}
}

func TestAllLevels(t *testing.T) {
	// test all the available logger levels.

	wh, buffer := setupBufferLogger(Debug, t)

	err := setupLogger([]BuildLogHandler{wh}, t)
	if err != nil {
		t.Errorf("There was an error while setting up the logger\n%s", err.Error())
		return
	}

	logger := Log()

	// all functions
	run := func(f func() string) bool {
		buffer.Reset()
		exp := f()
		res := strings.TrimSpace(buffer.String())

		if !strings.HasSuffix(res, exp) {
			t.Errorf("The buffer should be filled with %s", exp)
			return false
		}
		return true
	}

	debug := func() string {
		logger.Debug("some random text")
		return "DEBUG some random text"
	}

	debugf := func() string {
		logger.Debugf("a number that annoyed me [%d]", 666)
		return "DEBUG a number that annoyed me [666]"
	}

	// the info functions
	info := func() string {
		logger.Info("test")
		return "INFO test"
	}

	infof := func() string {
		logger.Infof("a string %d", 23)
		return "INFO a string 23"
	}

	warn := func() string {
		logger.Warn("some terrible message threatening mankind")
		return "WARNING some terrible message threatening mankind"
	}

	warnf := func() string {
		logger.Warnf("some terrible message threatening mankind: [%s]", "with an addition")
		return "WARNING some terrible message threatening mankind: [with an addition]"
	}

	erf := func() string {
		logger.Error("well I almost crashed... HELP")
		return "ERROR well I almost crashed... HELP"
	}

	errf := func() string {
		logger.Errorf("I am still alive, have some useless numbers: [0x%x]", 0x42)
		return "ERROR I am still alive, have some useless numbers: [0x42]"
	}

	for _, f := range [...]func() string{debug, debugf, info, infof, warn, warnf, erf, errf} {
		if !run(f) {
			return
		}
	}
}

func TestDoubleLoggingInit(t *testing.T) {
	// Attention this calles Setup already
	fh, err := setupTmpFile(Debug, t)
	if err != nil {
		t.Errorf("There should have not been an error here.\n%s", err.Error())
		return
	}

	err = setupLogger([]BuildLogHandler{fh}, t)
	if err != nil {
		t.Errorf("There was an error that happend while creating the logger.\n%s", err.Error())
		return
	}
	err = setupLogger([]BuildLogHandler{fh}, t)
	if err == nil {
		t.Error("The error for a double setup is missing.")
	}
}

func TestLoggingToFile(t *testing.T) {
	fh, err := setupTmpFile(Debug, t)
	if err != nil {
		t.Errorf("There was an error while setting up the file.\n%s", err.Error())
		return
	}

	err = setupLogger([]BuildLogHandler{fh}, t)
	if err != nil {
		t.Errorf("There was an error while setting up the logger.\n%s", err.Error())
		return
	}

	log := Log()
	log.Infof("the value %d", 42)

	data, err := os.ReadFile(fh.FileName())
	if err != nil {
		t.Errorf("Unexpected file error happend.\n%s", err.Error())
		return
	}
	str := strings.TrimSpace(string(data))
	if !strings.HasSuffix(str, "the value 42") {
		t.Errorf("an incorrect suffix was used, for line\n[%s]", str)
	}
}

func TestIssueOpeningFile(t *testing.T) {
	// so that the global variable will be returned to it's value after this
	// test.
	fh := NewFileHandler(Debug, "notExistingFile.log")
	fh.openFile = func() (*os.File, error) {
		return nil, errors.New("an error message from the os")
	}

	err := setupLogger([]BuildLogHandler{fh}, t)
	if err == nil {
		t.Error("Error should have been returned here")
	}
}
