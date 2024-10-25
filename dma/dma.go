package dma

import (
	"fmt"
	"time"

	"github.com/ghanithan/challenge2016/config"
	"github.com/ghanithan/challenge2016/instrumentation"
	loadcsv "github.com/ghanithan/challenge2016/loadCsv"
)

// Package to handle the Designated Market Area

// Defines the number of heirarchy used in representing DMA
// Here we have Country -> State -> City
const HEIRARCHY = 3

// Enum to represent the Heirarchy
const (
	CITY    = iota
	STATE   = iota
	COUNTRY = iota
)

// DMA is structured as follows
type Dma struct {
	Countries   map[string]Country
	UpdatedTime time.Time
}

type Country struct {
	Place
	States map[string]State
}

type State struct {
	Place
	Cities map[string]City
}

type City Place

type Place struct {
	Name string
	Code string
}

func fmtPlace(place Place) string {
	return fmt.Sprintf("%s (%s)", place.Name, place.Code)
}

func (place Place) String() string {
	return fmtPlace(place)
}

func (city City) String() string {
	return fmtPlace(Place(city))
}

// Utility to query and print the DMA
type QueryDma struct {
	CountryCode string `json:"cc"`
	StateCode   string `json:"stc"`
	CityCode    string `json:"ctyc"`
}

func (dma Dma) PrintDma(query QueryDma) {
	country := dma.Countries[query.CountryCode]
	state := country.States[query.StateCode]
	city := state.Cities[query.CityCode]
	fmt.Println(country)
	fmt.Println(state)
	fmt.Println(city)
}

type row []Place

func validateRow(slice []string) (row, error) {
	if slice != nil && len(slice) != HEIRARCHY*2 {
		return nil, fmt.Errorf("there is discrepency in the data loaded from CSV")
	} else {

		places := make([]Place, HEIRARCHY)
		for i := 0; i < HEIRARCHY; i++ {
			places[i].Code = slice[i]
			places[i].Name = slice[HEIRARCHY+i]
		}
		return places, nil
	}

}

func InitDma(config *config.Config, logger *instrumentation.GoLogger) (*Dma, error) {
	csv, err := loadcsv.LoadCsv(config.Data.FilePath)
	if err != nil {
		logger.Error("Error in InitDma: %s", err)
	}

	dma := Dma{}
	dma.Countries = make(map[string]Country)

	for _, row := range csv {
		places, err := validateRow(row)
		if err != nil {
			logger.Error("%s", err)
			return nil, err
		}
		if country, ok := dma.Countries[places[COUNTRY].Code]; ok {
			states := country.States
			if state, ok := states[places[STATE].Code]; ok {
				if present, ok := state.Cities[places[CITY].Code]; ok {
					error := "Duplicate rows are present"
					logger.Error(error, present)
					//return nil, fmt.Errorf("%s: %s", error, present)
				} else {
					state.Cities[places[CITY].Code] = City(places[CITY])
				}
			} else {
				state := State{
					Place:  places[STATE],
					Cities: make(map[string]City),
				}
				states[places[STATE].Code] = state
			}
		} else {
			country := Country{
				Place:  places[COUNTRY],
				States: make(map[string]State),
			}
			dma.Countries[places[COUNTRY].Code] = country
		}
	}
	dma.UpdatedTime = time.Now()
	return &dma, nil
}
