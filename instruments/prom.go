package instruments

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type PrometheusReporter struct {
	met Metric
}

func NewPrometheusReporter(met Metric) PrometheusReporter {
	return PrometheusReporter{
		met: met,
	}
}

func (m PrometheusReporter) Report(begin time.Time, method string, err error) {
	lvs := []string{"method", method, "error", fmt.Sprint(err != nil)}
	m.met.RequestCount.With(lvs...).Add(1)
	m.met.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
}

type Instrumenting interface {
	Report(time.Time, string, error)
}

type Metric struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
}

func PrometheusMetric(ns, svc string) Metric {
	fieldKeys := []string{"method", "error"}

	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: ns,
		Subsystem: svc,
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)

	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:  ns,
		Subsystem:  svc,
		Name:       "request_latency_microseconds",
		Help:       "Total duration of requests in seconds.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, fieldKeys)

	return Metric{
		RequestCount:   requestCount,
		RequestLatency: requestLatency,
	}
}

func DummyMetric(n string) Metric {
	return Metric{
		RequestCount:   generic.NewCounter(n + "_count"),
		RequestLatency: generic.NewHistogram(n+"_latency", 50),
	}
}
