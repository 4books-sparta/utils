package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIngestionServer(t *testing.T, status int, body string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/public/ingestion", r.URL.Path)
		assert.Equal(t, BasicAuth("pk", "sk"), r.Header.Get("Authorization"))
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("LANGFUSE_HOST", srv.URL)
	t.Setenv("LANGFUSE_PUBLIC_KEY", "pk")
	t.Setenv("LANGFUSE_SECRET_KEY", "sk")
	return New()
}

func TestIngestionMultiStatusWithoutErrorsIsSuccess(t *testing.T) {
	c := newIngestionServer(t, http.StatusMultiStatus, `{"successes":[{"id":"a","status":201}],"errors":[]}`)
	res := IngestionResponse{}
	require.NoError(t, c.Ingestion(context.Background(), &Ingestion{}, &res))
	assert.Equal(t, http.StatusMultiStatus, res.Code)
	assert.Len(t, res.Successes, 1)
}

func TestIngestionMultiStatusWithErrorsFails(t *testing.T) {
	c := newIngestionServer(t, http.StatusMultiStatus, `{"successes":[],"errors":[{"id":"a","status":400,"message":"bad score","error":"x"}]}`)
	res := IngestionResponse{}
	err := c.Ingestion(context.Background(), &Ingestion{}, &res)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad score")
	assert.Len(t, res.Errors, 1)
}

func TestIngestionUnauthorizedFails(t *testing.T) {
	c := newIngestionServer(t, http.StatusUnauthorized, `{"message":"Unauthorized"}`)
	err := c.Ingestion(context.Background(), &Ingestion{}, &IngestionResponse{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unauthorized")
}
