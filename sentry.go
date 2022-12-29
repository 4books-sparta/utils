package utils

import (
	"github.com/getsentry/raven-go"
)

type SentryReporter struct {
	client *raven.Client
}

func NewSentryReporter(dsn string) (*SentryReporter, error) {
	c, err := raven.New(dsn)
	if err != nil {
		return nil, err
	}

	return &SentryReporter{
		client: c,
	}, nil
}

func (r *SentryReporter) Report(err error, ctx ...string) {
	tags := map[string]string{}

	var key string
	for _, val := range ctx {
		if key == "" {
			key = val
			continue
		}

		tags[key] = val
		key = ""
	}

	r.client.CaptureError(err, tags)
}

func (r *SentryReporter) Message(message string, ctx ...string) {
	tags := map[string]string{}

	var key string
	for _, val := range ctx {
		if key == "" {
			key = val
			continue
		}

		tags[key] = val
		key = ""
	}

	r.client.CaptureMessage(message, tags)
}
