package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ghanithan/challenge2016/config"
	"github.com/ghanithan/challenge2016/dma"
	"github.com/ghanithan/challenge2016/handlers"
	"github.com/ghanithan/challenge2016/instrumentation"
	"github.com/gorilla/mux"
)

type Server struct {
	HttpService *http.Server
	AppService  *handlers.Service
}

func InitServer(config *config.Config, dmaService *dma.Dma, logger *instrumentation.GoLogger) (Server, context.CancelFunc) {
	c := context.WithValue(context.Background(), "version", "0.0.1")
	c, cancel := context.WithCancel(c)
	router := mux.NewRouter()

	appService := handlers.Service{
		Context:    c,
		DmaService: dmaService,
		Logger:     logger,
	}

	router = appService.AddHanlders(router)

	httpService := http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%s", config.HttpServer.Port),
	}

	return Server{
		HttpService: &httpService,
		AppService:  &appService,
	}, cancel
}
