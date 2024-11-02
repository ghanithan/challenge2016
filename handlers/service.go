package handlers

import (
	"context"

	"github.com/ghanithan/challenge2016/dma"
	"github.com/ghanithan/challenge2016/instrumentation"
	"github.com/gorilla/mux"
)

type Service struct {
	Context    context.Context
	DmaService *dma.Dma
	Logger     *instrumentation.GoLogger
}

func (app *Service) AddHanlders(router *mux.Router) *mux.Router {
	router.Handle("/version", app.GetVersion()).Methods("GET")
	router.Handle("/list/places", app.GetListPlaces()).Methods("GET")
	router.Handle("/distributor", app.PostAddNewDistributor()).Methods("POST")
	router.Handle("/distributor", app.GetDistributor()).Methods("GET")
	router.Handle("/distributor", app.DeleteDistributor()).Methods("DELETE")
	router.Handle("/distributor", app.PatchUpdateDistributor()).Methods("PATCH")
	router.Handle("/permission/{distributor}", app.GetDistributorPermission()).Methods("GET")

	return router
}
