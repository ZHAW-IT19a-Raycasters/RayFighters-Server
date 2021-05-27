package controller

import (
	"time"

	"github.zhaw.ch/raycasters/server/gamemap"
)

// mapTimeMeta is a struct to save a MapMeta a Time since when it was live
type mapTimeMeta struct {
	meta  *gamemap.MapMeta
	since time.Time
}

// newMapMeta creates a new mapTimeMeta element
// as a Parameter is uses the MapMeta and sets the since time to now
func newMapMeta(meta *gamemap.MapMeta) mapTimeMeta {
	return mapTimeMeta{
		meta:  meta,
		since: time.Now(),
	}
}
