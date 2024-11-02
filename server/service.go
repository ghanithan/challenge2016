package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ghanithan/challenge2016/config"
	"github.com/ghanithan/challenge2016/dma"
	"github.com/ghanithan/challenge2016/handlers"
	"github.com/ghanithan/challenge2016/instrumentation"
	ghandlers "github.com/gorilla/handlers"
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
	// Catch all OPTIONS requests and handle CORS preflight
	router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept-Encoding")
		w.WriteHeader(http.StatusOK)
	})
	router.Use(corsMiddleware)
	router.Use(ghandlers.CompressHandler)

	router = appService.AddHanlders(router)

	httpService := http.Server{
		Handler: router,
		Addr:    fmt.Sprintf("%s:%s", config.HttpServer.Host, config.HttpServer.Port),
	}

	return Server{
		HttpService: &httpService,
		AppService:  &appService,
	}, cancel
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept-Encoding")

		// Proceed with the request
		next.ServeHTTP(w, r)
	})
}
