package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.zhaw.ch/raycasters/server/log"
	"github.zhaw.ch/raycasters/server/setup"
	"github.zhaw.ch/raycasters/server/web"
)

// logFile is the file, where to the logs shall be logged to
const logFile string = "logs.txt"

// the system logger
var logger = log.Log()

func setupLogger() error {
	b := log.NewBuilder()

	// register file handler
	fh := log.NewFileHandler(log.Info, logFile)
	b.RegisterHandler(fh)

	// register console handler
	ch := log.NewConsoleHandler(log.Debug)
	b.RegisterHandler(ch)

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

	server, err := web.NewServer(&conf.Web)
	if err != nil {
		logger.Errorf("unable to start the server due to [%s]", err.Error())
		return
	}

	done := make(chan struct{})
	ctx := context.Background()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)

		select {
		case sig := <-sigs:
			fmt.Println(sig)
		case <-done:
			return
		}

		server.ServerStop(ctx)
		close(done)
	}()

	err = server.Start()
	if !errors.Is(err, http.ErrServerClosed) {
		close(done)
		fmt.Println(err)
	}

	logger.Info("awaiting server shutdown")

	<-done
	logger.Info("Shutdown completed")
}
