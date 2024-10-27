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

	qubeDma.PrintPlaces()

	queryDma := dma.QueryDma{
		CountryCode: "BE",
		StateCode:   "WLG",
	}

	// qubeDma.PrintDma(queryDma)

	include := []dma.QueryDma{}
	exclude := []dma.QueryDma{}

	include = append(include, queryDma)

	distributor, err := qubeDma.AddDistributor("distributor1", nil)
	if err != nil {
		logger.Error("%s", err)
	}

	qubeDma.IncludeDistributorPermission(distributor, include, exclude, logger)

	logger.Info("distributor1 added")
	qubeDma.PrintDistributors()

	qubeDma.PrintPlacesFrom(queryDma)

	queryDma = dma.QueryDma{
		CountryCode: "IN",
		StateCode:   "TN",
	}

	exclude = append(exclude, queryDma)

	distributor, err = qubeDma.AddDistributor("distributor2", distributor)
	if err != nil {
		logger.Error("%s", err)
	}
	err = qubeDma.IncludeDistributorPermission(distributor, include, exclude, logger)
	if err != nil {
		logger.Error("%s", err)
	} else {
		logger.Info("distributor2 added")
		qubeDma.PrintPlacesFrom(queryDma)

	}

	include = include[0:0]

	distributor, err = qubeDma.AddDistributor("distributor3", nil)
	if err != nil {
		logger.Error("%s", err)
	}

	err = qubeDma.IncludeDistributorPermission(distributor, include, exclude, logger)
	if err != nil {
		logger.Error("%s", err)
	}
	queryDma = dma.QueryDma{
		CountryCode: "BE",
	}

	qubeDma.PrintPlacesFrom(queryDma)

}
