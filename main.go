package main

import (
	"fmt"
	"os"
	"os/signal"

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

	done := make(chan struct{})
	ctx := web.ShutdownContext()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)

		var sig os.Signal

		select {
		case sigs <- sig:
			fmt.Println()
			fmt.Println(sig)
		case <-done:
			return
		}

		// TODO: global channel shutdown
		server.ServerStop(ctx)
		close(done)
	}()

	err = server.Start()
	if err != nil {
		done <- struct{}{}
		close(done)
		return
	}
	fmt.Println("awaiting server shutdown")

	<-done
	fmt.Println("Shutdown completed")
}
