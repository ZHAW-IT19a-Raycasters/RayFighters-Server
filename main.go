package main

import (
	"fmt"

	"github.zhaw.ch/raycasters/server/log"
	"github.zhaw.ch/raycasters/server/setup"
)

func setupLogger() error {
	b := log.NewBuilder()

	// TODO:

	// register file handler

	// register console handler

	return b.Build()
}

func main() {
	setUp := setup.NewSetup()
	conf := setUp.Configuration()

	if conf == nil {
		return
	}

	err := setupLogger()

	if err != nil {
		return
	}

	fmt.Println("WebConfig : ", conf.Web.String())
}
