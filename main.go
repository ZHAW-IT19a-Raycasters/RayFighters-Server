package main

import (
	"github.zhaw.ch/raycasters/server/utils"
)

func main() {
	conf := utils.Setup()
	if conf == nil {
		return
	}
}
