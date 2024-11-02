package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ghanithan/challenge2016/config"
	"github.com/ghanithan/challenge2016/dma"
	"github.com/ghanithan/challenge2016/instrumentation"
	server "github.com/ghanithan/challenge2016/server"
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

	// distributor, err := qubeDma.AddDistributor("distributor1", nil)
	// if err != nil {
	// 	logger.Error("%s", err)
	// }

	// include := []string{"IN"}
	// exclude := []string{}

	// qubeDma.IncludeDistributorPermission(distributor, include, exclude, logger)

	// qubeDma.PrintPlacesFrom("IN-TN-CENAI")

	service, cancel := server.InitServer(config, qubeDma, &logger)
	defer cancel()

	// Start the server
	go func() {
		logger.Info("Server listening on ", logger.String("addr", service.HttpService.Addr))
		if err := service.HttpService.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server...")
	cancel()

}
