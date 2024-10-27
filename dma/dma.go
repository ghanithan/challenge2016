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
	CITY = iota
	STATE
	COUNTRY
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
	distributors *Distributor
}

type Lookup struct {
	places       map[string]*Place
	distributors map[string]*Distributor
}

type Tier int

type Place struct {
	id            uuid.UUID
	Type          Tier //Enum
	Tag           string
	Name          string
	Code          string
	rightsOwnedBy *Distributor
	next          []*Place
	up            *Place
}

func fmtPlace(place Place) string {
	return fmt.Sprintf("%s (%s)", place.Name, place.Code)
}

func (place Place) getType() int {
	return int(place.Type)
}

func (place Place) setType(tier int) {
	place.Type = Tier(tier)
}

func (tier Tier) String() string {
	str := ""
	switch int(tier) {
	case CITY:
		str = "City"
	case STATE:
		str = "State"
	case COUNTRY:
		str = "Country"
	default:
		str = "City"
	}
	return str
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

func (dma *Dma) queryToPlace(query string) (*Place, error) {
	dma.RLock()
	defer dma.RUnlock()
	place, ok := dma.lookup.places[query]
	if !ok {
		return nil, fmt.Errorf("queried place is not supported")
	}
	return place, nil
}

func (dma *Dma) PrintDma(query string) {
	dma.RLock()
	defer dma.RUnlock()

	place := dma.lookup.places[query]
	printDmaInternal(place)
	fmt.Println("Place Id:", place.id)
	fmt.Println("Distributor:\n", place.rightsOwnedBy)

}

func printDmaInternal(place *Place) {
	if place == nil {
		return
	}
	printDmaInternal(place.up)
	fmt.Printf("%s: %s", place.Type, place)
}

type row []Place

func validateRow(slice []string) (row, error) {
	if slice != nil && len(slice) != HEIRARCHY*2 {
		return nil, fmt.Errorf("there is discrepency in the data loaded from CSV")
	} else {

		makeTag := func(index int, slice []string) string {
			tag := ""
			for i := HEIRARCHY - 1; i > index; i-- {
				tag = fmt.Sprintf("%s%s", tag, fmt.Sprintf("%s"))
			}
			if len(tag) == 0 {
				tag = slice[index]
			}
			return tag
		}

		places := make([]Place, HEIRARCHY)
		for i := 0; i < HEIRARCHY; i++ {
			places[i].id = uuid.New()
			places[i].Code = slice[i]
			places[i].Name = slice[HEIRARCHY+i]
			places[i].Type = Tier(HEIRARCHY)
			places[i].Tag = fmt.Sprintf()
		}
		pl

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
	dma.lookup.places = make(map[string]*Place)
	dma.lookup.distributors = make(map[string]*Distributor)
	world := Place{
		id:   uuid.New(),
		Name: "World",
		Code: "World",
	}
	dma.data.places = &world
	dma.data.distributors = &Distributor{}

	for _, row := range csv {
		places, err := validateRow(row)
		if err != nil {
			logger.Error("%s", err)
			return nil, err
		}
		if country, ok := dma.lookup.places[places[COUNTRY].Code]; ok {

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

type Distributor struct {
	id       uuid.UUID
	name     string
	includes []*Place
	excludes []*Place
	prev     *Distributor
	next     []*Distributor
}

func (dist Distributor) String() string {
	return fmt.Sprintf("%s: %s\n - Include: %q\n -Exclude %q\n", dist.id, dist.name, dist.includes, dist.excludes)
}

func (dma *Dma) PrintDistributors() {
	dma.RLock()
	defer dma.RUnlock()
	printDistrbibutorsInternal(dma.data.distributors, 0)
}

func printDistrbibutorsInternal(node *Distributor, stage int) {
	if node == nil {
		return
	}

	fmt.Println(stage, node)

	for _, child := range node.next {
		printDistrbibutorsInternal(child, stage+1)
	}

}

func (dma *Dma) GetDistributor(name string) (*Distributor, error) {
	if dist, ok := dma.lookup.distributors[name]; ok {
		return dist.entity, nil
	} else {
		return nil, fmt.Errorf("distributor not found in the list")
	}
}

func (dma *Dma) AddDistributor(name string, parent *Distributor) (*Distributor, error) {
	dma.Lock()
	defer dma.Unlock()
	if existingDistributor, ok := dma.lookup.distributors[name]; ok {
		return existingDistributor.entity, fmt.Errorf("distributor already present in the list")
	}
	dist := &Distributor{
		id:   uuid.New(),
		name: name,
	}
	if parent == nil {
		parent = dma.data.distributors
	}
	parent.next = append(parent.next, dist)
	dist.prev = parent

	dma.lookup.distributors[name] = distributorWrapper{
		entity: dist,
	}
	return dist, nil
}

func (place *Place) isDistributorPresent() bool {
	return place.rightsOwnedBy != nil
}

func (place *Place) isDistributor(dist *Distributor) bool {
	return place.rightsOwnedBy == dist
}

func (dma *Dma) queryDmaToPlaces(queries []QueryDma) ([]*Place, error) {
	places := []*Place{}
	for _, query := range queries {
		place, err := dma.queryToPlace(query)
		if err != nil {
			return []*Place{}, err
		}
		places = append(places, place)

	}
	return places, nil
}

// func (dma *Dma) traversePlaces(place *Place, placeList []*Place) error {
// 	dma.RLock()
// 	defer dma.RUnlock()

// 	return traversePlacesInternal(place, placeList)
// }

// func traversePlacesInternal(node *Place, placeList []*Place) error {
// 	if node == nil {
// 		return nil
// 	}

// 	for _, child := range node.next {
// 		traversePlacesInternal(child, placeList)
// 	}
// }

// func (dma *Dma) placePartOfExcludes(place *Place, excludes []QueryDma) error {
// 	places, err := dma.queryDmaToPlaces(excludes)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	for _, excludePlace := range places {

// 	}
// }

// func checkConflictDistributor(place *Place, dist *Distributor, excludes []*Place) error {
// 	if place == nil {
// 		return nil
// 	}

// 	if !place.isDistributor(dist.prev) {
// 		if placePartOfExcludes(place, excludes) {

// 		}
// 		return fmt.Errorf("%s",
// 			"the permission cannot be issued since parent distributor has no rights in that place")
// 	}

// 	for _, child := range place.next {
// 		err := checkConflictDistributor(child, dist)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func checkConflictDistributor(place *Place, dist *Distributor) error {
// 	if place == nil {
// 		return nil
// 	}

// 	if !place.isDistributor(dist.prev) {
// 		return fmt.Errorf("%s",
// 			"the permission cannot be issued since parent distributor has no rights in that place")
// 	}

// 	for _, child := range place.next {
// 		err := checkConflictDistributor(child, dist)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func assignDistributor(place *Place, dist *Distributor) {
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

func excludeDistributor(place *Place, dist *Distributor) {
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

func (dma *Dma) appendDistributorInclude(distributor *Distributor, place *Place,
	logger instrumentation.GoLogger) error {
	defer logger.TimeTheFunction(time.Now(), "appendDistributorInclude")
	dma.Lock()
	defer dma.Unlock()

	if place.isDistributorPresent() {
		return fmt.Errorf("%s: %s in %s", distributor.name, distributorAlreadyPresentError, place)
	} else {
		assignDistributor(place, distributor)
	}
	temp := dma.lookup.distributors[distributor.name]
	temp.entity.includes = append(temp.entity.includes,
		place)
	dma.lookup.distributors[distributor.name] = temp
	return nil
}

func (dma *Dma) appendDistributorExclude(distributor *Distributor, place *Place, logger instrumentation.GoLogger) error {
	defer logger.TimeTheFunction(time.Now(), "appendDistributorExclude")
	dma.Lock()
	defer dma.Unlock()

	if place.isDistributor(distributor) {
		excludeDistributor(place, distributor)
	} else {
		logger.Info("%s: %s in %s", distributor.name, distributorAlreadyExcludedError, place)
	}
	temp := dma.lookup.distributors[distributor.name]
	temp.entity.excludes = append(temp.entity.excludes,
		place)
	dma.lookup.distributors[distributor.name] = temp
	return nil
}

func (dma *Dma) IncludeDistributorPermission(distributor *Distributor, includes []QueryDma, excludes []QueryDma,
	logger instrumentation.GoLogger) error {
	defer logger.TimeTheFunction(time.Now(), "IncludeDistributorPermission")

	logger.Info("adding inclusions")
	for _, include := range includes {
		place, err := dma.queryToPlace(include)
		if err != nil {
			return err
		}
		err = dma.appendDistributorInclude(distributor, place, logger)
		if err != nil {
			return err
		}
	}

	logger.Info("adding exclusions")
	for _, exclude := range excludes {
		place, err := dma.queryToPlace(exclude)
		if err != nil {
			return err
		}
		err = dma.appendDistributorExclude(distributor, place, logger)
		if err != nil {
			return err
		}
	}

	return nil

}
