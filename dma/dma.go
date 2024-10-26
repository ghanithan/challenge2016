package dma

import (
	"fmt"
	"sync"
	"time"

	"github.com/ghanithan/challenge2016/config"
	"github.com/ghanithan/challenge2016/instrumentation"
	loadcsv "github.com/ghanithan/challenge2016/loadCsv"
	"github.com/google/uuid"
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

const (
	distributorAlreadyPresentError = "distributor is already present for the place"
)

// DMA is structured as follows
type Dma struct {
	mu          sync.Mutex
	data        Data
	lookup      Lookup
	updatedTime time.Time
}

type Data struct {
	places       *Place
	distributors *distributor
}

type Lookup struct {
	countries    map[string]Country
	distributors map[string]distributorWrapper
}

type Country struct {
	Place  *Place
	States map[string]State
}

type State struct {
	Place  *Place
	Cities map[string]City
}

type City struct {
	Place *Place
}

type Place struct {
	id            uuid.UUID
	Name          string
	Code          string
	rightsOwnedBy *distributor
	next          []*Place
	up            *Place
}

func fmtPlace(place Place) string {
	return fmt.Sprintf("%s (%s)", place.Name, place.Code)
}

func (place Place) String() string {
	return fmtPlace(place)
}

// Utility to query and print the DMA
type QueryDma struct {
	CountryCode string `json:"cc"`
	StateCode   string `json:"stc,omitempty"`
	CityCode    string `json:"ctyc,omitempty"`
}

func (dma Dma) PrintDma(query QueryDma) {
	if len(query.CountryCode) == 0 {
		fmt.Println("The query must have country code")
		return
	}
	country := dma.lookup.countries[query.CountryCode]
	fmt.Println("Country:", country.Place)
	if len(query.StateCode) == 0 {
		fmt.Println("Place Id:", country.Place.id)
		return
	}
	state := country.States[query.StateCode]
	fmt.Println("State:", state.Place)

	if len(query.CityCode) == 0 {
		fmt.Println("Place Id:", state.Place.id)
		return
	}
	city := state.Cities[query.CityCode]
	fmt.Println("City:", city.Place)
	fmt.Println("Place Id:", city.Place.id)

}

type row []Place

func validateRow(slice []string) (row, error) {
	if slice != nil && len(slice) != HEIRARCHY*2 {
		return nil, fmt.Errorf("there is discrepency in the data loaded from CSV")
	} else {

		places := make([]Place, HEIRARCHY)
		for i := 0; i < HEIRARCHY; i++ {
			places[i].id = uuid.New()
			places[i].Code = slice[i]
			places[i].Name = slice[HEIRARCHY+i]
		}

		return places, nil
	}

}

func (place *Place) addUp(up *Place) {
	place.up = up
}

func (place *Place) addNext(next *Place) {
	place.next = append(place.next, next)
}

func InitDma(config *config.Config, logger *instrumentation.GoLogger) (*Dma, error) {
	csv, err := loadcsv.LoadCsv(config.Data.FilePath)
	if err != nil {
		logger.Error("Error in InitDma: %s", err)
	}

	dma := Dma{}
	dma.lookup.countries = make(map[string]Country)
	world := Place{
		id:   uuid.New(),
		Name: "World",
		Code: "World",
	}
	dma.data.places = &world

	for _, row := range csv {
		places, err := validateRow(row)
		if err != nil {
			logger.Error("%s", err)
			return nil, err
		}
		if country, ok := dma.lookup.countries[places[COUNTRY].Code]; ok {
			states := country.States
			if state, ok := states[places[STATE].Code]; ok {
				if present, ok := state.Cities[places[CITY].Code]; ok {
					error := "Duplicate rows are present"
					logger.Error(error, present)
					//return nil, fmt.Errorf("%s: %s", error, present)
				} else {
					place := &places[CITY]
					place.addUp(state.Place)
					state.Place.addNext(place)
					state.Cities[places[CITY].Code] = City{
						Place: place,
					}

				}
			} else {
				place := &places[STATE]
				place.addUp(country.Place)
				country.Place.addNext(place)
				state := State{
					Place:  place,
					Cities: make(map[string]City),
				}

				// Add city
				place = &places[CITY]
				place.addUp(state.Place)
				state.Place.addNext(place)
				state.Cities[places[CITY].Code] = City{
					Place: place,
				}

				states[places[STATE].Code] = state
			}
		} else {
			place := &places[COUNTRY]
			place.up = &world
			world.next = append(world.next, place)
			country := Country{
				Place:  place,
				States: make(map[string]State),
			}
			dma.lookup.countries[places[COUNTRY].Code] = country

			place = &places[STATE]
			place.addUp(country.Place)
			country.Place.addNext(place)
			state := State{
				Place:  place,
				Cities: make(map[string]City),
			}
			country.States[places[STATE].Code] = state

			// Add city
			place = &places[CITY]
			place.addUp(state.Place)
			state.Place.addNext(place)
			state.Cities[places[CITY].Code] = City{
				Place: place,
			}

		}
	}
	dma.updatedTime = time.Now()
	return &dma, nil
}

func (dma Dma) PrintTree() {
	printTreeInternal(dma.data.places, 0)
}

func printTreeInternal(node *Place, stage int) {
	if node == nil {
		return
	}

	fmt.Println(stage, node)
	fmt.Printf("%q\n", node.next)
	for _, child := range node.next {
		printTreeInternal(child, stage+1)
	}
}

// Distributor Datastructure

// I am looking to have a tight coupling between DMA and Disbributor Datastructures
// This should help us retrieve information at a time complexity if O(1)

type distributor struct {
	id       uuid.UUID
	name     string
	includes []*Place
	excludes []*Place
	prev     *distributor
	next     *distributor
}

type distributorWrapper struct {
	entity *distributor
}

func (dma *Dma) AddDistributor(name string) (*distributor, error) {
	if existingDistributor, ok := dma.lookup.distributors[name]; ok {
		return existingDistributor.entity, fmt.Errorf("distributor already present in the list")
	}
	distributor := distributor{
		id:   uuid.New(),
		name: name,
	}
	dma.lookup.distributors[name] = distributorWrapper{
		entity: &distributor,
	}
	return &distributor, nil
}

func (place Place) isDistributorPresent() bool {
	return place.rightsOwnedBy != nil

}

func (dma *Dma) fetchDistributor(name string) *distributor {
	return dma.lookup.distributors[name].entity
}

func (dma *Dma) appendDistributorInclude(name string, place *Place) error {
	distributor := dma.fetchDistributor(name)
	if place.isDistributorPresent() {
		return fmt.Errorf("%s in %s", distributorAlreadyPresentError, place)
	} else {
		place.rightsOwnedBy = distributor
	}
	temp := dma.lookup.distributors[name]
	temp.entity.includes = append(temp.entity.includes,
		place)
	dma.lookup.distributors[name] = temp
	return nil
}

func (dma *Dma) IncludeDistributorPermission(name string, includes []QueryDma, excludes []QueryDma,
	logger instrumentation.GoLogger) error {

	_, err := dma.AddDistributor(name)
	if err != nil {
		logger.Info("%s", err)
	}

	for _, include := range includes {
		var place *Place
		switch {
		case len(include.CityCode) != 0:
			place = dma.lookup.countries[include.CountryCode].States[include.StateCode].Cities[include.CityCode].Place
		case len(include.StateCode) != 0:
			place = dma.lookup.countries[include.CountryCode].States[include.StateCode].Place
		case len(include.CountryCode) != 0:
			place = dma.lookup.countries[include.CountryCode].Place
		}
		dma.appendDistributorInclude(name, place)
	}

	return nil

}
