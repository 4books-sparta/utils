package --service-name--

import (
	"github.com/4books-sparta/utils"
)

type reportingMiddleware struct {
	reporter utils.ErrorReporter
	next     Service
}

func ReportingMiddleware(reporter utils.ErrorReporter) Middleware {
	return func(next Service) Service {
		return reportingMiddleware{reporter, next}
	}
}

