package handlers

import (
	"encoding/json"
	"io"
	"net/http"
)

type AddNewDistributorRequest struct {
	Name    string   `json:"name"`
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

func (service *Service) PostAddNewDistributor() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "invalid body")
		}
		defer r.Body.Close()

		request := AddNewDistributorRequest{}

		err = json.Unmarshal(body, &request)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "invalid json")
		}

		dist, err := service.DmaService.AddDistributor(request.Name, nil)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "")
		}

		err = service.DmaService.IncludeDistributorPermission(dist, request.Include, request.Exclude, *service.Logger)
		if err != nil {
			service.Logger.Error(err.Error())
			FailureResponse(w, http.StatusInternalServerError, "")
		}
	})
}
