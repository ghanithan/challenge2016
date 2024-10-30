package dma

import (
	"fmt"
	"strings"
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

func (place *Place) setType(tier int) {
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
	return fmt.Sprintf("%s (%s) - %s", place, place.id, place.rightsOwnedBy)
}

func (place Place) String() string {
	return fmtPlace(place)
}

func (place *Place) AddChildNode(childNode *Place) {
	place.addNext(childNode)
	childNode.addUp(place)
}

func (place *Place) AddParentNode(parentNode *Place) {
	parentNode.addNext(place)
	place.addUp(parentNode)
}

// Utility to query and print the DMA
type QueryDma struct {
	CountryCode string `json:"cc"`
	StateCode   string `json:"stc,omitempty"`
	CityCode    string `json:"ctyc,omitempty"`
}

func (query QueryDma) String() string {
	output := ""
	switch {
	case query.CityCode == "" && query.StateCode == "" && query.CountryCode == "":
		break
	case query.CityCode == "" && query.StateCode == "":
		output = query.CountryCode
	case query.CityCode == "":
		output = fmt.Sprintf("%s-%s", query.CountryCode, query.StateCode)
	default:
		output = fmt.Sprintf("%s-%s-%s", query.CountryCode, query.StateCode, query.CityCode)
	}
	return output
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

	if place, ok := dma.lookup.places[query]; ok {
		printDmaInternal(place)
		fmt.Println("Place Id:", place.id)
		fmt.Println("Distributor:\n", place.rightsOwnedBy)
	} else {
		fmt.Println(query, "not found")
	}

}

func printDmaInternal(place *Place) {
	if place == nil {
		return
	}
	printDmaInternal(place.up)
	fmt.Printf("%s: %s\n", place.Type, place)
}

func validateRow(slice []string) (*Place, error) {
	if slice != nil && len(slice) != HEIRARCHY*2 {
		return nil, fmt.Errorf("there is discrepency in the data loaded from CSV")
	} else {

		makeTag := func(index int, slice []string) string {
			tag := slice[HEIRARCHY-1]
			for i := HEIRARCHY - 2; i >= index; i-- {
				tag = fmt.Sprintf("%s-%s", tag, slice[i])
			}
			if len(tag) == 0 {
				tag = slice[index]
			}
			return tag
		}

		var leaf *Place
		var root *Place
		for i := 0; i < HEIRARCHY; i++ {
			place := Place{}
			place.id = uuid.New()
			place.Code = slice[i]
			place.Name = slice[HEIRARCHY+i]
			place.setType(i)
			place.Tag = makeTag(i, slice)
			if leaf == nil {
				leaf = &place
				root = leaf
			} else {
				root.addUp(&place)
				root = &place
			}

		}
		// if slice[1] == "WLG" {
		// 	fmt.Println("WLG here")
		// 	fmt.Println(leaf.getType())
		// }
		return leaf, nil
	}

}

func (place *Place) addUp(up *Place) {
	place.up = up
}

func (place *Place) addNext(next *Place) {
	place.next = append(place.next, next)
}

func populateData(dma *Dma, leaf *Place, logger *instrumentation.GoLogger) *Place {

	if leaf == nil {
		return dma.data.places
	}

	if present, ok := dma.lookup.places[leaf.Tag]; ok {

		return present
	}
	parent := populateData(dma, leaf.up, logger)
	dma.Lock()
	defer dma.Unlock()
	parent.AddChildNode(leaf)
	dma.lookup.places[leaf.Tag] = leaf
	return leaf
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
		place, err := validateRow(row)
		if err != nil {
			logger.Error("%s", err)
			return nil, err
		}
		populateData(&dma, place, logger)

	}

	dma.updatedTime = time.Now()
	if dma.data.places == nil {
		return nil, fmt.Errorf("the initialization failed")
	}
	return &dma, nil
}

func (dma *Dma) PrintPlaces() {
	dma.RLock()
	defer dma.RUnlock()

	printPlacesInternal(dma.data.places, 3)
}

func (dma *Dma) PrintPlacesLookup() {
	dma.RLock()
	defer dma.RUnlock()
	for tag, place := range dma.lookup.places {
		fmt.Println(tag, ":", place)
	}
}

func (dma *Dma) PrintPlacesFrom(query string) {
	dma.RLock()
	defer dma.RUnlock()
	place, err := dma.queryToPlace(query)
	if err != nil {
		fmt.Println(err)
	}
	printPlacesInternal(place, place.getType())
}

func (dma *Dma) GetPlaceByCode(query string) *Place {
	dma.RLock()
	defer dma.RUnlock()

	return dma.lookup.places[query]
}

func (dma *Dma) GetDistributorByName(name string) *Distributor {
	dma.RLock()
	defer dma.RUnlock()

	return dma.lookup.distributors[name]
}

func printPlacesInternal(node *Place, stage int) {
	if node == nil {
		return
	}
	fmt.Println(strings.Repeat("\t", HEIRARCHY-stage), Tier(stage), node.fmtPlaceWithRights())
	// fmt.Printf("%s\n", node.next)
	for _, child := range node.next {
		printPlacesInternal(child, stage-1)
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
	up       *Distributor
	next     []*Distributor
}

func (dist *Distributor) String() string {
	if dist == nil {
		return "No distributor"
	}
	return fmt.Sprintf("%s (%s)", dist.name, dist.id)
}

func (dist Distributor) PrintDistributorDetails() string {
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

	fmt.Println(stage, node.PrintDistributorDetails())

	for _, child := range node.next {
		printDistrbibutorsInternal(child, stage+1)
	}

}

func (dma *Dma) GetDistributor(name string) (*Distributor, error) {
	if dist, ok := dma.lookup.distributors[name]; ok {
		return dist, nil
	} else {
		return nil, fmt.Errorf("distributor not found in the list")
	}
}

func (dma *Dma) AddDistributor(name string, parent *Distributor) (*Distributor, error) {
	dma.Lock()
	defer dma.Unlock()
	if existingDistributor, ok := dma.lookup.distributors[name]; ok {
		return existingDistributor, fmt.Errorf("distributor already present in the list")
	}
	dist := &Distributor{
		id:   uuid.New(),
		name: name,
	}
	if parent == nil {
		parent = dma.data.distributors
	}
	parent.next = append(parent.next, dist)
	dist.up = parent

	dma.lookup.distributors[name] = dist
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
		place, err := dma.queryToPlace(fmt.Sprint(query))
		if err != nil {
			return []*Place{}, err
		}
		places = append(places, place)

	}
	return places, nil
}

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
	temp.includes = append(temp.includes,
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
		logger.Info(distributor.name, distributorAlreadyExcludedError, place)
	}
	temp := dma.lookup.distributors[distributor.name]
	temp.excludes = append(temp.excludes,
		place)
	dma.lookup.distributors[distributor.name] = temp
	return nil
}

func (dma *Dma) IncludeDistributorPermission(distributor *Distributor, includes []QueryDma, excludes []QueryDma,
	logger instrumentation.GoLogger) error {

	defer logger.TimeTheFunction(time.Now(), "IncludeDistributorPermission")

	err := dma.CheckConflictBeforeChange(distributor, includes, excludes, logger)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	logger.Info("adding inclusions")
	for _, include := range includes {
		place, err := dma.queryToPlace(fmt.Sprint(include))
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
		place, err := dma.queryToPlace(fmt.Sprint(exclude))
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

func (dma *Dma) CheckConflictBeforeChange(distributor *Distributor, includes []QueryDma, excludes []QueryDma,
	logger instrumentation.GoLogger) error {

	defer logger.TimeTheFunction(time.Now(), "CheckConflictBeforeChange")

	logger.Info("fetching inclusions")
	inclusionPlaces := make(map[string]*Place)
	for _, include := range includes {
		place, err := dma.queryToPlace(fmt.Sprint(include))
		if err != nil {
			return err
		}
		inclusionPlaces[place.id.String()] = place
	}
	for _, place := range distributor.includes {
		inclusionPlaces[place.id.String()] = place
	}
	logger.Info(fmt.Sprintf("%q", inclusionPlaces))

	logger.Info("fetching exclusions")
	exclusionPlaces := make(map[string]*Place)
	for _, exclude := range excludes {
		place, err := dma.queryToPlace(fmt.Sprint(exclude))
		if err != nil {
			return err
		}
		exclusionPlaces[place.id.String()] = place
	}
	for _, place := range distributor.excludes {
		exclusionPlaces[place.id.String()] = place
	}

	logger.Info(fmt.Sprintf("%q", exclusionPlaces))

	err := CheckConflictDistributor(dma, distributor, inclusionPlaces, exclusionPlaces)
	if err != nil {
		return err
	}
	return nil
}

func CheckConflictDistributor(dma *Dma, dist *Distributor, includes map[string]*Place, exlcudes map[string]*Place) error {
	for _, child := range includes {
		err := checkConflictDistributor(dma, child, dist, exlcudes)
		if err != nil {
			return err
		}
	}

	return nil

}

func checkConflictDistributor(dma *Dma, node *Place, dist *Distributor, exlcudes map[string]*Place) error {
	if node == nil {
		return nil
	}
	//inclusion check
	if node.rightsOwnedBy != dist.up {
		if _, ok := exlcudes[node.id.String()]; ok {
			return nil
		} else {
			return fmt.Errorf("parent(%s) lacks the rights to add the distributor", dist.up)
		}
	}

	for _, child := range node.next {
		err := checkConflictDistributor(dma, child, dist, exlcudes)
		if err != nil {
			return err
		}
	}

	return nil

}
