package requests

import (
	"context"
	"net/http"
	"strconv"

	"goji.io/pat"
)

func DecodeGetByIntIdRequest(_ context.Context, req *http.Request) (interface{}, error) {
	var request IdRequest

	id, err := strconv.Atoi(pat.Param(req, "id"))
	if err != nil {
		return nil, err
	}

	request.Id = id

	return request, nil
}

func DecodeEmptyRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return struct{}{}, nil
}

func DecodeRefreshableRequest(_ context.Context, req *http.Request) (interface{}, error) {
	_, ok := req.URL.Query()["refresh"]
	return ok, nil
}
