package utils

import (
	"time"

	"github.com/go-kit/kit/metrics"
	kitPrometheus "github.com/go-kit/kit/metrics/prometheus"
	stdPrometheus "github.com/prometheus/client_golang/prometheus"
)

type Instrumenting interface {
	Report(time.Time, string, error)
}

type Metric struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
}

func PrometheusMetric(name, svc string) Metric {
	fieldKeys := []string{"method", "error"}

	requestCount := kitPrometheus.NewCounterFrom(stdPrometheus.CounterOpts{
		Namespace: name,
		Subsystem: svc,
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)

	requestLatency := kitPrometheus.NewSummaryFrom(stdPrometheus.SummaryOpts{
		Namespace: name,
		Subsystem: svc,
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	return Metric{
		RequestCount:   requestCount,
		RequestLatency: requestLatency,
	}
}
