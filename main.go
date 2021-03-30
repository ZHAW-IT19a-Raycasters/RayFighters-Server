package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.zhaw.ch/raycasters/server/log"
	"github.zhaw.ch/raycasters/server/setup"
	"github.zhaw.ch/raycasters/server/web"
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

		select {
		case sig := <-sigs:
			fmt.Println(sig)
		case <-done:
			return
		}
		// TODO: global channel shutdown
		server.ServerStop(ctx)
		close(done)
	}()

	err = server.Start()
	if !errors.Is(err, http.ErrServerClosed) {
		close(done)
		fmt.Println(err)
	}

	fmt.Println("awaiting server shutdown")

	<-done
	fmt.Println("Shutdown completed")
}
