package main

import (
	"fmt"

	server "github.com/ghanithan/challenge2016/Server"
	"github.com/ghanithan/challenge2016/config"
	"github.com/ghanithan/challenge2016/dma"
	"github.com/ghanithan/challenge2016/instrumentation"
)

func main() {
	fmt.Println("Starting Challenge 2016...")

	logger := instrumentation.InitInstruments()

	config, err := config.GetConfig(logger, "./setting/default.yaml")

	if err != nil {
		logger.Error(err.Error())
	}

	qubeDma, err := dma.InitDma(config, &logger)
	if err != nil {
		logger.Error(err.Error())
	}

	service, cancel := server.InitServer(config, qubeDma, &logger)
	defer cancel()

	logger.Info("Server listening on :", logger.String("port", service.HttpService.Addr))

	err = service.HttpService.ListenAndServe()
	if err != nil {
		logger.Error("error in initializing the server", err)
	}

	logger.Info("Server listening on port:", service.HttpService.Addr)

}
