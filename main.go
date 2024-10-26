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

	qubeDma.IncludeDistributorPermission("distributor1", include, exclude, logger)

	logger.Info("distributor1 added")
	qubeDma.PrintDistributors()

	qubeDma.PrintDma(queryDma)

	queryDma = dma.QueryDma{
		CountryCode: "BE",
		StateCode:   "WLG",
		CityCode:    "BUIGE",
	}

	exclude = append(exclude, queryDma)


	qubeDma.IncludeDistributorPermission("distributor2", include, exclude, logger)

	queryDma = dma.QueryDma{
		CountryCode: "BE",
	}

	qubeDma.PrintPlacesFrom(queryDma)

}
