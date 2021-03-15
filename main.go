package main

import (
	"fmt"

	"github.zhaw.ch/raycasters/server/utils"
)

func main() {
	conf := utils.Setup()
	if conf == nil {
		return
	}

	fmt.Println("WebConfig : ", conf.Web.String())
}
