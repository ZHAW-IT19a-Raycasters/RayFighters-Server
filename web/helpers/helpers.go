package helpers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

// RoomIdParameter is the name used to identify
// the room variable.
const RoomIdParameter string = "roomID"

// RoomIdFailedId is the id returned in case of an error.
const RoomIdFailedId int = 0

// ErrNoValidRoomId is the error returned in case of
// an inability to parse the url
type ErrNoValidRoomId struct {
	addr *url.URL
}

// Error is the error interface
func (err ErrNoValidRoomId) Error() string {
	return fmt.Sprintf("no valid roomID in URL: %v", err.addr)
}

// Is satisfies the errors.Is scheme
func (err ErrNoValidRoomId) Is(target error) bool {
	_, ok := target.(ErrNoValidRoomId)
	return ok
}

// ErrUnableToParseRoomId is the error returned in case
// that the id was not a valid id.
type ErrUnableToParseRoomId struct {
	id       string
	internal error
}

// Error is the error interface
func (err ErrUnableToParseRoomId) Error() string {
	return fmt.Sprintf("couldn't parse roomID [%s] from URL: [%s]", err.id, err.internal.Error())
}

// Is satisfies the errors.Is scheme
func (err ErrUnableToParseRoomId) Is(target error) bool {
	_, ok := target.(ErrUnableToParseRoomId)
	return ok
}

// ParseRoomID parses the RoomID from the request URL.
// It will return an error and zero if no valid roomID
// was found in URL or if the string conversion failed.
func ParseRoomID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	if len(vars[RoomIdParameter]) == 0 {
		return RoomIdFailedId, ErrNoValidRoomId{
			r.URL,
		}
	}
	roomIDStr := vars[RoomIdParameter]
	roomID, err := strconv.Atoi(roomIDStr)

	if err != nil {
		return RoomIdFailedId, ErrUnableToParseRoomId{
			id:       roomIDStr,
			internal: err,
		}
	}

	return roomID, nil
}
