package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

func (service *Service) DeleteDistributor() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query().Get("id")
		if len(query) == 0 {
			query = r.URL.Query().Get("name")
		}

		if len(query) != 0 {
			values := strings.Split(query, ",")
			for _, value := range values {
				err := service.DmaService.DeleteDistributor(value)
				if err != nil {
					service.Logger.Error(err.Error())
					FailureResponse(w, http.StatusNotFound, fmt.Sprintf("distributor %s is not found", value))
					return
				}
			}

		} else {
			FailureResponse(w, http.StatusUnprocessableEntity, "query parameter 'name' missing")
		}

		SuccessResponse(w, nil)

	})
}
