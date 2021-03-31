package helpers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// parse roomID from URL.
// returns an error if no vaild roomID in URL or if string conversion failed.
func ParseRoomID(r *http.Request) (int, error){
	vars := mux.Vars(r)
	if len(vars["roomID"])==0 {
		return -1, fmt.Errorf("no valid roomID in URL: %v",r.URL)
	}
	roomIDStr := vars["roomID"]
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		return -1, fmt.Errorf("couldn't parse roomID from URL: %v", err)
	}

	return roomID, nil
}