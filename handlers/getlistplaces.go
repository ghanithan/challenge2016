package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ghanithan/challenge2016/dma"
)

type ListPlaces struct {
	Data []*dma.Place `json:"data"`
}

func (service *Service) GetListPlaces() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := ListPlaces{}
		if queryIn := r.URL.Query().Get("in"); len(queryIn) != 0 {
			// Adding support to filter multiple places in the same request
			queries := strings.Split(queryIn, ",")
			for _, query := range queries {
				queryResult, err := service.DmaService.GetPlaceByCode(query)
				if err != nil {
					FailureResponse(w, http.StatusNotFound, err.Error())
					return
				}
				data.Data = append(data.Data, queryResult)
			}
		} else {
			// List all places in the dma
			data.Data = service.DmaService.GetPlaces()
		}

		if dataJson, err := json.Marshal(data); err != nil {
			service.Logger.Error("Failed to send places list ", service.Logger.String("host", r.Host))
			FailureResponse(w, http.StatusNotFound, "Not Found")
		} else {
			SuccessResponse(w, dataJson)
		}

	})
}