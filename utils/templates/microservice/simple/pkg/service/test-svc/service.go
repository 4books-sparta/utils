package --service-name--

import (
	"github.com/4books-sparta/utils"
	"github.com/go-redis/redis/v8"

	"--module--/pkg/service/helper"
)

type Service interface {
}

type microService struct {
	repo   Repo
	rc     redis.UniversalClient
	helper helper.Service
	metric utils.Metric
}

func New(repo Repo, rep utils.ErrorReporter, met utils.Metric, rc redis.UniversalClient) Service {
	var svc Service
	{
		svc = microService{
			repo:   repo,
			rc:     rc,
			helper: helper.New(repo, rep, met, rc),
			metric: met,
		}
		if rc != nil {
			svc = CachingMiddleware(rc, rep)(svc)
		}
		svc = ReportingMiddleware(rep)(svc)
		svc = InstrumentingMiddleware(met)(svc)
	}

	return svc
}

