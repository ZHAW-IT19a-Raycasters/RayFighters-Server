package controller

import (
	"time"

	"github.zhaw.ch/raycasters/server/gamemap"
)

type mapTimeMeta struct {
	meta  *gamemap.MapMeta
	since time.Time
}

func newMapMeta(meta *gamemap.MapMeta) mapTimeMeta {
	return mapTimeMeta{
		meta:  meta,
		since: time.Now(),
	}
}
