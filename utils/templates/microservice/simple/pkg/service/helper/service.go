package helper

import (
	"github.com/4books-sparta/utils"
	"github.com/go-redis/redis/v8"
)

type Service interface {
}

func New(repo Repo, rep utils.ErrorReporter, met utils.Metric, rc redis.UniversalClient) Service {
	var svc Service
	return svc
}
