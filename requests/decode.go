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

func DecodeQueryPagedRequest(_ context.Context, req *http.Request) (interface{}, error) {
	var request PagedRequest

	if limit, ok := req.URL.Query()["limit"]; ok && len(limit) > 0 {
		if v, err := strconv.Atoi(limit[0]); err == nil {
			request.Limit = uint(v)
		}
	}
	if offset, ok := req.URL.Query()["offset"]; ok && len(offset) > 0 {
		if v, err := strconv.Atoi(offset[0]); err == nil {
			request.Offset = uint(v)
		}
	}

	return request, nil
}

func DecodeEmptyRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return struct{}{}, nil
}

func DecodeRefreshableRequest(_ context.Context, req *http.Request) (interface{}, error) {
	_, ok := req.URL.Query()["refresh"]
	return ok, nil
}

func CanBeDecoded(req *http.Request) bool {
	if req.Method == http.MethodPatch {
		return true
	}
	if req.Method == http.MethodPost {
		return true
	}
	if req.Method == http.MethodPut {
		return true
	}

	return false
}
