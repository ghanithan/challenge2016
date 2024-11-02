package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ghanithan/challenge2016/dma"
	"github.com/gorilla/mux"
)

type GetDistributorPermissionResponse struct {
	Place     Place `json:"place"`
	Permitted bool  `json:"permitted"`
}

type Place struct {
	Name string `json:"name"`
	Id   string `json:"id"`
	Tag  string `json:"tag"`
}

func dmaPlacetoPlace(place *dma.Place) Place {
	return Place{
		Name: place.String(),
		Id:   place.Id.String(),
		Tag:  place.Tag,
	}
}

func (service *Service) GetDistributorPermission() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pathParam, ok := vars["distributor"]
		if !ok {
			FailureResponse(w, http.StatusBadRequest, "the distributor path paramter is not present. '/permission/{distirbutorId}?in=placeTag1,placeTag2")
			return
		}

		distributor, err := service.DmaService.GetDistributor(pathParam)
		if err != nil {
			FailureResponse(w, http.StatusNotFound, err.Error())
			return
		}
		response := []GetDistributorPermissionResponse{}
		if queryIn := r.URL.Query().Get("in"); len(queryIn) != 0 {
			// Adding support to filter multiple places in the same request
			queries := strings.Split(queryIn, ",")
			for _, query := range queries {
				queryResult, err := service.DmaService.GetPlaceByCode(query)
				if err != nil {
					FailureResponse(w, http.StatusNotFound, err.Error())
					return
				}

				responseValue := GetDistributorPermissionResponse{
					Place:     dmaPlacetoPlace(queryResult),
					Permitted: queryResult.RightsOwnedBy == distributor,
				}
				response = append(response, responseValue)
			}

		} else {
			FailureResponse(w, http.StatusBadRequest, "the distributor query paramter is not present. '/permission/{distirbutorId}?in=placeTag1,placeTag2")
			return
		}

		if dataJson, err := json.Marshal(response); err != nil {
			service.Logger.Error("Failed to send premission list ", service.Logger.String("host", r.Host))
			FailureResponse(w, http.StatusNotFound, "Not Found")
		} else {
			SuccessResponse(w, dataJson)
		}

	})
}
