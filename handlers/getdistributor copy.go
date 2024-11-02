package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ghanithan/challenge2016/dma"
)

type GetDistributorResponse struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Includes []string `json:"includedPlaces"`
	Excludes []string `json:"excludedPlaces"`
}

func (service *Service) GetDistributor() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response []byte
		var err error

		// both query by id and query by name is supported
		// Highest priority goes to id
		query := r.URL.Query().Get("id")
		if len(query) == 0 {
			query = r.URL.Query().Get("name")
		}
		if len(query) != 0 {
			dists := []GetDistributorResponse{}
			values := strings.Split(query, ",")
			for _, value := range values {
				var dist *dma.Distributor
				dist, err = service.DmaService.GetDistributor(value)
				if err != nil {
					service.Logger.Error(err.Error())
					FailureResponse(w, http.StatusNotFound, fmt.Sprintf("distributor %s is not found", value))
					return
				}
				distributor := GetDistributorResponse{
					Includes: dist.GetIncludesAsTags(),
					Excludes: dist.GetExcludesAsTags(),
					Id:       string(dist.Id.String()),
					Name:     dist.Name,
				}
				dists = append(dists, distributor)
			}

			response, err = json.Marshal(dists)

		} else {
			dists := service.DmaService.GetDistributors()
			response, err = json.Marshal(dists)
		}

		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "")
			return
		}
		SuccessResponse(w, response)

	})
}
