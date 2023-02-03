package --service-name--

import (
	"fmt"
	"time"

	"github.com/4books-sparta/utils"
)

type instrumentingMiddleware struct {
	met  utils.Metric
	next Service
}

func InstrumentingMiddleware(met utils.Metric) Middleware {
	return func(next Service) Service {
		return instrumentingMiddleware{met, next}
	}
}

func (m instrumentingMiddleware) report(begin time.Time, method string, err error) {
	lvs := []string{"method", method, "error", fmt.Sprint(err != nil)}
	m.met.RequestCount.With(lvs...).Add(1)
	m.met.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
}

