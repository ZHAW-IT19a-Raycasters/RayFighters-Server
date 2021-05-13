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
// const logFile string = "logs.txt"

// the system logger
var logger = log.Log()

// setupLogger will configure the logger
func setupLogger(conf setup.LoggerConfig) error {
	b := log.NewBuilder()

	// register file handler

	// this error is impossible to get as incorrect
	// configuration should be stoped before
	// the programm arrives here
	lv, _ := log.LevelDetect(conf.File.Level)

	fh := log.NewFileHandler(lv, conf.File.FilePath)
	err := b.RegisterHandler(fh)

	// there might have been an error while
	// opening the file
	if err != nil {
		return err
	}

	// register console handler
	lv, _ = log.LevelDetect(conf.Console.Level)
	ch := log.NewConsoleHandler(lv)

	// new console handler will never return an error
	// on registraction.
	_ = b.RegisterHandler(ch)

	return b.Build()
}

func main() {
	setUp := setup.NewSetup()
	conf := setUp.Configuration()

	if conf == nil {
		return
	}

	err := setupLogger(conf.Logger)

	if err != nil {
		fmt.Println(err.Error())
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
