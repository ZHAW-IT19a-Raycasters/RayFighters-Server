package main

import (
	"fmt"

	"github.zhaw.ch/raycasters/server/log"
<<<<<<< HEAD
	"github.zhaw.ch/raycasters/server/setup"
=======
	"github.zhaw.ch/raycasters/server/utils"
	"github.zhaw.ch/raycasters/server/web"
>>>>>>> Implement router
)

func main() {
	setUp := setup.NewSetup()
	conf := setUp.Configuration()

	if conf == nil {
		return
	}

	err := log.Setup(log.Debug)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("WebConfig : ", conf.Web.String())
	web.ServingRoute(&conf.Web)
}
