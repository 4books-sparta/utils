package api

import (
	"context"
	"encoding/base64"
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
			"Authorization": basicAuth(publicKey, secretKey),
		},
	}

	return &Client{
		restClient: &restClient,
	}
}

func (c *Client) Ingestion(_ context.Context, req *Ingestion, res *IngestionResponse) error {
	ep, _ := req.Path()
	return c.restClient.Post(ep, req, res)
}

func basicAuth(publicKey, secretKey string) string {
	auth := publicKey + ":" + secretKey
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
