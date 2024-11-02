package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ghanithan/challenge2016/dma"
)

type UpdateDistributorRequest struct {
	Name    string `json:"name"`
	Include Change `json:"include"`
	Exclude Change `json:"exclude"`
}

type Change struct {
	Add    []string `json:"add"`
	Delete []string `json:"delete"`
}

func (service *Service) PatchUpdateDistributor() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "invalid body")
			return
		}
		defer r.Body.Close()

		request := UpdateDistributorRequest{}

		err = json.Unmarshal(body, &request)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "invalid json")
			return
		}

		var distributor *dma.Distributor

		distributor, err = service.DmaService.GetDistributor(request.Name)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusNotFound, fmt.Sprintf("distributor '%s' not found", request.Name))
			return
		}

		err = service.DmaService.IncludeDistributorPermission(distributor, request.Include.Add,
			request.Exclude.Add, *service.Logger)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError,
				fmt.Sprintf("could not update the distributor with the inclusions: %s", err))
			return
		}

		err = service.DmaService.DeleteDistributorInclude(distributor, request.Include.Delete, *service.Logger)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError,
				fmt.Sprintf("could not update the distributor with deletions: %s", err))
			return
		}

		err = service.DmaService.DeleteDistributorExclude(distributor, request.Exclude.Delete, *service.Logger)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError,
				fmt.Sprintf("could not update the distributor with deletions: %s", err))
			return
		}

		jsonDist, err := json.Marshal(distributor)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "")
			return
		}

		SuccessResponse(w, jsonDist)
	})
}
