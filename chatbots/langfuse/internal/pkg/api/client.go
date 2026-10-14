package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/4books-sparta/utils"
)

const (
	langfuseDefaultEndpoint = "https://cloud.langfuse.com"
)

type Client struct {
	restClient *utils.MicroserviceClient
}

func New() *Client {
	lfHost := os.Getenv("LANGFUSE_HOST")
	if lfHost == "" {
		lfHost = langfuseDefaultEndpoint
	}

	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")

	restClient := utils.MicroserviceClient{
		TimeOut: 5 * time.Second,
		Url:     lfHost,
		Port:    0,
		PermanentHeaders: map[string]string{
			"Authorization": BasicAuth(publicKey, secretKey),
		},
	}

	return &Client{
		restClient: &restClient,
	}
}

// Ingestion posts a batch to the legacy ingestion endpoint. Langfuse answers
// 207 Multi-Status with per-event outcomes, so the status alone is not an
// error: the body is decoded into res and the first failed event, if any, is
// returned as the error.
func (c *Client) Ingestion(_ context.Context, req *Ingestion, res *IngestionResponse) error {
	ep, _ := req.Path()
	rc := *c.restClient
	rc.ErrorParser = func(r *http.Response) error { return ingestionResult(r, res) }
	return rc.Post(ep, req, res)
}

func ingestionResult(r *http.Response, res *IngestionResponse) error {
	if r.StatusCode != http.StatusMultiStatus {
		return fmt.Errorf("langfuse ingestion: %w", utils.ServiceResponse2Error(r))
	}
	if res == nil {
		res = &IngestionResponse{}
	}
	if err := res.Decode(r.Body); err != nil {
		return fmt.Errorf("langfuse ingestion: decoding %d response: %w", r.StatusCode, err)
	}
	res.Code = r.StatusCode
	if len(res.Errors) > 0 {
		e := res.Errors[0]
		return fmt.Errorf("langfuse ingestion: event %s rejected (%d): %s %s", e.ID, e.Status, e.Message, e.Error)
	}
	return nil
}

// BasicAuth returns the HTTP Basic Authorization header value for a Langfuse key pair.
func BasicAuth(publicKey, secretKey string) string {
	auth := publicKey + ":" + secretKey
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
