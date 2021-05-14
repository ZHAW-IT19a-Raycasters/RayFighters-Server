package log

import (
	"strings"

	"errors"
	"testing"
)

func TestDetectLevel(t *testing.T) {
	const is = "dEbug  "
	lv, err := LevelDetect(is)

	if err != nil {
		t.Error("the input value is valid, so there should not be an error here")
		return
	} else if lv != Debug {
		t.Errorf("incorrect level returend got [%s], should be [%s]", lv.String(), Debug.String())
		return
	}

	const isNot = "ifo"
	_, err = LevelDetect(isNot)
	if err == nil {
		t.Error("there should be an error here on an invalid input")
	} else if !errors.Is(err, ErrUnknownLevel{}) {
		t.Error("incorrect error type detected")
	} else if !strings.Contains(err.Error(), "unknown level name") {
		t.Error("wrong error string used")
	}

}
