package main

import (
	"fmt"
	"os"

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

	server := web.NewServer(&conf.Web)
	server.Start()
	sigs := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	go func() {
		defer server.Stop()
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- struct{}{}
	}()
	fmt.Println("awaiting signal")
	<-done
}
