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
	distributorAlreadyPresentError  = "distributor is already present for the place"
	distributorAlreadyExcludedError = "distributor is already excluded for the place"
)

// DMA is structured as follows
type Dma struct {
	sync.RWMutex
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

func (place Place) fmtPlaceWithRights() string {
	return fmt.Sprintf("%s (%s) - %s", place.Name, place.Code, place.rightsOwnedBy)
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

func (dma *Dma) queryToPlace(query QueryDma) (*Place, error) {
	dma.RLock()
	defer dma.RUnlock()
	var place *Place
	switch {
	case len(query.CityCode) != 0:
		place = dma.lookup.countries[query.CountryCode].States[query.StateCode].Cities[query.CityCode].Place
	case len(query.StateCode) != 0:
		place = dma.lookup.countries[query.CountryCode].States[query.StateCode].Place
	case len(query.CountryCode) != 0:
		place = dma.lookup.countries[query.CountryCode].Place
	}
	if place == nil {
		return nil, fmt.Errorf("queried place is not supported")
	}
	return place, nil
}

func (dma *Dma) PrintDma(query QueryDma) {
	dma.RLock()
	defer dma.RUnlock()
	if len(query.CountryCode) == 0 {
		fmt.Println("The query must have country code")
		return
	}
	country := dma.lookup.countries[query.CountryCode]
	fmt.Println("Country:", country.Place)
	if len(query.StateCode) == 0 {
		fmt.Println("Place Id:", country.Place.id)
		fmt.Println("Distributor:\n", country.Place.rightsOwnedBy)
		return
	}
	state := country.States[query.StateCode]
	fmt.Println("State:", state.Place)

	if len(query.CityCode) == 0 {
		fmt.Println("Place Id:", state.Place.id)
		fmt.Println("Distributor:\n", state.Place.rightsOwnedBy)
		return
	}
	city := state.Cities[query.CityCode]
	fmt.Println("City:", city.Place)
	fmt.Println("Place Id:", city.Place.id)
	fmt.Println("Distributor:\n", city.Place.rightsOwnedBy)

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
	dma.lookup.distributors = make(map[string]distributorWrapper)
	world := Place{
		id:   uuid.New(),
		Name: "World",
		Code: "World",
	}
	dma.data.places = &world
	dma.data.distributors = &distributor{}

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

func (dma *Dma) PrintPlaces() {
	dma.RLock()
	defer dma.RUnlock()

	printPlacesInternal(dma.data.places, 0)
}

func (dma *Dma) PrintPlacesFrom(query QueryDma) {
	dma.RLock()
	defer dma.RUnlock()
	place, err := dma.queryToPlace(query)
	if err != nil {
		fmt.Println(err)
	}
	printPlacesInternal(place, 0)
}

func printPlacesInternal(node *Place, stage int) {
	if node == nil {
		return
	}

	fmt.Println(stage, node.fmtPlaceWithRights())
	fmt.Printf("%q\n", node.next)
	for _, child := range node.next {
		printPlacesInternal(child, stage+1)
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

func (dist distributor) String() string {
	return fmt.Sprintf("%s: %s\n - Include: %q\n -Exclude %q\n", dist.id, dist.name, dist.includes, dist.excludes)
}

func (dma *Dma) PrintDistributors() {
	dma.RLock()
	defer dma.RUnlock()

}

func printDistrbibutorsInternal(node *distributor, stage int) {
	if node == nil {
		return
	}

	fmt.Println(stage, node)

	printDistrbibutorsInternal(node.next, stage+1)

}

func (dma *Dma) AddDistributor(name string) (*distributor, error) {
	dma.Lock()
	defer dma.Unlock()
	if existingDistributor, ok := dma.lookup.distributors[name]; ok {
		return existingDistributor.entity, fmt.Errorf("distributor already present in the list")
	}
	dist := &distributor{
		id:   uuid.New(),
		name: name,
	}
	dist.next = dma.data.distributors
	temp := dma.data.distributors
	dma.data.distributors = dist
	temp.prev = dist

	dma.lookup.distributors[name] = distributorWrapper{
		entity: dist,
	}
	return dist, nil
}

func (place *Place) isDistributorPresent() bool {
	return place.rightsOwnedBy != nil
}

func (place *Place) isDistributor(dist *distributor) bool {
	return place.rightsOwnedBy == dist
}

func checkConflictDistributor(place *Place, dist *distributor) error {
	if place == nil {
		return nil
	}

	if place.isDistributor(dist) {
		return fmt.Errorf("%s", distributorAlreadyPresentError)
	}

	for _, child := range place.next {
		err := checkConflictDistributor(child, dist)
		if err != nil {
			return err
		}
	}
	return nil
}

func assignDistributor(place *Place, dist *distributor) {
	if place == nil {
		return
	}

	place.rightsOwnedBy = dist

	for _, child := range place.next {
		assignDistributor(child, dist)
	}
}

func (place *Place) removeDistributor() {
	place.rightsOwnedBy = nil
}

func excludeDistributor(place *Place, dist *distributor) {
	if place == nil {
		return
	}

	if place.isDistributor(dist) {
		place.removeDistributor()
	}

	for _, child := range place.next {
		excludeDistributor(child, dist)
	}
}

func (dma *Dma) fetchDistributor(name string) *distributor {
	return dma.lookup.distributors[name].entity
}

func (dma *Dma) appendDistributorInclude(name string, place *Place, logger instrumentation.GoLogger) error {
	defer logger.TimeTheFunction(time.Now(), "appendDistributorInclude")
	dma.Lock()
	defer dma.Unlock()

	distributor := dma.fetchDistributor(name)
	if place.isDistributorPresent() {
		return fmt.Errorf("%s in %s", distributorAlreadyPresentError, place)
	} else {
		err := checkConflictDistributor(place, distributor)
		if err != nil {
			return err
		}
		assignDistributor(place, distributor)
	}
	temp := dma.lookup.distributors[name]
	temp.entity.includes = append(temp.entity.includes,
		place)
	dma.lookup.distributors[name] = temp
	return nil
}

func (dma *Dma) appendDistributorExclude(name string, place *Place) error {
	dma.Lock()
	defer dma.Unlock()

	distributor := dma.fetchDistributor(name)
	if place.isDistributor(distributor) {
		excludeDistributor(place, distributor)
	} else {
		return fmt.Errorf("%s in %s", distributorAlreadyExcludedError, place)
	}
	temp := dma.lookup.distributors[name]
	temp.entity.excludes = append(temp.entity.excludes,
		place)
	dma.lookup.distributors[name] = temp
	return nil
}

func (dma *Dma) IncludeDistributorPermission(name string, includes []QueryDma, excludes []QueryDma,
	logger instrumentation.GoLogger) error {
	defer logger.TimeTheFunction(time.Now(), "IncludeDistributorPermission")
	logger.Info("adding districution", name)

	_, err := dma.AddDistributor(name)
	if err != nil {
		logger.Info("%s", err)
	}

	logger.Info("adding inclusions")
	for _, include := range includes {
		place, err := dma.queryToPlace(include)
		if err != nil {
			return err
		}
		dma.appendDistributorInclude(name, place, logger)
	}

	logger.Info("adding exclusions")
	for _, exclude := range excludes {
		place, err := dma.queryToPlace(exclude)
		if err != nil {
			return err
		}
		dma.appendDistributorExclude(name, place)
	}

	return nil

}
