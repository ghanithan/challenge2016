package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ghanithan/challenge2016/dma"
)

type AddNewDistributorRequest struct {
	Name    string   `json:"name"`
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
	Parent  string   `json:"parent"`
}

func (service *Service) PostAddNewDistributor() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "invalid body")
			return
		}
		defer r.Body.Close()

		request := AddNewDistributorRequest{}

		err = json.Unmarshal(body, &request)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "invalid json")
			return
		}

		var parent *dma.Distributor
		if len(request.Parent) != 0 {
			parent, err = service.DmaService.GetDistributor(request.Parent)
			if err != nil {
				service.Logger.Error(err.Error())
				FailureResponse(w, http.StatusNotFound, fmt.Sprintf("parent '%s' not found", request.Parent))
				return
			}
		}
		fmt.Printf("here %q\n", parent)
		dist, err := service.DmaService.AddDistributor(request.Name, parent)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "")
			return
		}

		err = service.DmaService.IncludeDistributorPermission(dist, request.Include, request.Exclude, *service.Logger)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError,
				fmt.Sprintf("could not add the distributor: %s", err))
			if err = service.DmaService.DeleteDistributor(request.Name); err != nil {
				service.Logger.Error(err.Error())
			}
			return
		}

		jsonDist, err := json.Marshal(dist)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "")
			return
		}

		SuccessResponse(w, jsonDist)
	})
}
