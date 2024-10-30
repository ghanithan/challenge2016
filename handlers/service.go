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
	router.Handle("/version", app.GetVersion(router))
	return router
}
