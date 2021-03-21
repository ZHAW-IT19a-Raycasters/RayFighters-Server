package main

import (
	"fmt"

	"github.zhaw.ch/raycasters/server/log"
	"github.zhaw.ch/raycasters/server/setup"
)

func main() {
	conf := setup.Setup()
	if conf == nil {
		return
	}

	err := log.Setup(log.Debug)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("WebConfig : ", conf.Web.String())
}
