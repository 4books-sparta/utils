package utils

import (
	"context"
	"crypto/subtle"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ProbeResponse struct {
	Status uint8 `json:"status"`
}

func MakeProbeEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return ProbeResponse{
			Status: 1,
		}, nil
	}
}

func MakePrometheusHandler(u, p, r string) http.HandlerFunc {
	handler := promhttp.Handler()

	return func(w http.ResponseWriter, req *http.Request) {
		user, pass, ok := req.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(u)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(p)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+r+`"`)
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		handler.ServeHTTP(w, req)
	}
}
