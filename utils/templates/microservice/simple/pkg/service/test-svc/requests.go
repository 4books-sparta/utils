package --service-name--

import (
	"context"
	"net/http"

	"github.com/4books-sparta/utils"

)

func DecodeSimpleRequest(_ context.Context, req *http.Request) (interface{}, error) {
	var request struct{}
//	id, err := strconv.Atoi(pat.Param(req, "id"))

	err := utils.DecodeRequestBody(req, &request)
	if err != nil {
		return nil, err
	}

	err = utils.ValidateRequest(request)
	if err != nil {
		return nil, err
	}

	return request, nil
}
