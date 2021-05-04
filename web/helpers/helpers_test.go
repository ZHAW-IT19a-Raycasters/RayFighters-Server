package helpers_test

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.zhaw.ch/raycasters/server/web/helpers"
)

func TestParseRoomID(t *testing.T) {
	const id = 0

	r, _ := http.NewRequest("GET", "/", nil)

	vars := map[string]string{
		helpers.RoomIdParameter: fmt.Sprintf("%d", id),
	}
	r = mux.SetURLVars(r, vars)

	rId, err := helpers.ParseRoomID(r)

	if err != nil {
		t.Errorf("There should not be an error here [%s]", err.Error())
		return
	}

	if rId != id {
		t.Error("incorrect id was returned")
	}
}

func TestParseRoomNoId(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)

	vars := map[string]string{}
	r = mux.SetURLVars(r, vars)

	rId, err := helpers.ParseRoomID(r)

	if err == nil {
		t.Errorf("There should be an error here")
		return
	} else if !errors.Is(err, helpers.ErrNoValidRoomId{}) {
		t.Error("incorrect error returned")
		return
	} else if !strings.Contains(err.Error(), "no valid roomID") {
		t.Error("false substring returned")
	}

	if rId != helpers.RoomIdFailedId {
		t.Error("incorrect id returned")
	}
}

func TestParseRoomNoInteger(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)

	vars := map[string]string{
		helpers.RoomIdParameter: "ThisIsNoID",
	}
	r = mux.SetURLVars(r, vars)

	rId, err := helpers.ParseRoomID(r)

	if err == nil {
		t.Errorf("There should be an error here")
		return
	} else if !errors.Is(err, helpers.ErrUnableToParseRoomId{}) {
		t.Error("incorrect error returned")
		return
	} else if !strings.Contains(err.Error(), "couldn't parse roomID") {
		t.Error("false substring returned")
	}

	if rId != helpers.RoomIdFailedId {
		t.Error("incorrect id returned")
	}
}
