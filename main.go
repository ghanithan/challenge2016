package main

import (
	"fmt"

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
	//logger.Info("%v", qubeDma)

	qubeDma.PrintTree()

	queryDma := dma.QueryDma{
		CountryCode: "BE",
		StateCode:   "WLG",
		CityCode:    "BUIGE",
	}

	qubeDma.PrintDma(queryDma)

}
