package handlers

import (
	"encoding/json"
	"net/http"
)

type Version struct {
	Version string `json:"version"`
}

func (service *Service) GetVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := Version{
			Version: service.Context.Value("version").(string),
		}

		if versionJson, err := json.Marshal(version); err != nil {
			service.Logger.Error("Failed to send version information", service.Logger.String("host", r.Host))
			FailureResponse(w, "Not Found")
		} else {
			SuccessResponse(w, versionJson)
		}

	})
}
