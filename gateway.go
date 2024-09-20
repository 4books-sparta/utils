package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/4books-sparta/utils/cache"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kitHttp "github.com/go-kit/kit/transport/http"
)

const (
	AuthCtxKey = "req-auth"
)

type Forwarder struct {
	auth Authorizer
}

type Authorization struct {
	User  string
	Role  string
	Error error
}

type Authorizer interface {
	Authorize(context.Context, *http.Request) Authorization
}

func NewForwarder(af Authorizer) *Forwarder {
	return &Forwarder{
		auth: af,
	}
}

func (f *Forwarder) forward(e endpoint.Endpoint, dec kitHttp.DecodeRequestFunc, opts ...kitHttp.ServerOption) *kitHttp.Server {
	mid := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(errorEncoder),
		kitHttp.ServerBefore(blockWrongSlug),
		kitHttp.ServerBefore(plugRefresh),
		kitHttp.ServerAfter(writeCORS),
	}
	mid = append(mid, opts...)

	return kitHttp.NewServer(e, dec, encodeResponse, mid...)
}

func (f *Forwarder) Forward(e endpoint.Endpoint, dec kitHttp.DecodeRequestFunc) *kitHttp.Server {
	return f.forward(e, dec)
}

func (f *Forwarder) SecureForward(e endpoint.Endpoint, dec kitHttp.DecodeRequestFunc) *kitHttp.Server {
	return f.forward(secureWrap(e, true), dec, kitHttp.ServerBefore(f.plugAuth))
}

func plugRefresh(ctx context.Context, req *http.Request) context.Context {
	_, refresh := req.URL.Query()["refresh"]
	if refresh {
		ctx = cache.GetContextWithForceRefreshCache(ctx, true)
	}
	return ctx
}

func blockWrongSlug(ctx context.Context, req *http.Request) context.Context {
	slug, ok := req.URL.Query()["slug"]
	if ok && len(slug) > 0 {
		//Check slug and fill error
		if !IsSlug(slug[0]) {
			return context.WithValue(ctx, "auth", errors.New("wrong-slug"))
		}
	}
	return ctx
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	setCORS(w)
	w.Header().Set("Content-Type", "application/json")

	c, ok := err.(WithCode)
	if ok {
		w.WriteHeader(c.Code())
	} else {
		w.WriteHeader(400)
	}

	body := map[string]string{
		"message": err.Error(),
	}
	_ = json.NewEncoder(w).Encode(body)
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	if v, ok := response.(RedirectResponse); ok {
		w.Header().Set("Location", v.RedirectTo())
		w.WriteHeader(http.StatusTemporaryRedirect)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func writeCORS(ctx context.Context, w http.ResponseWriter) context.Context {
	setCORS(w)

	return ctx
}

func Preflight(w http.ResponseWriter, _ *http.Request) {
	setCORS(w)
	w.WriteHeader(200)
}

func setCORS(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	h.Set("Access-Control-Allow-Headers", "Content-Type, Accept-Language, Authorization, X-Cms-Version, x-client-type, x-client-version")
}

type RedirectResponse interface {
	RedirectTo() string
}

type WithCode interface {
	Code() int
}

func (f *Forwarder) plugAuth(ctx context.Context, req *http.Request) context.Context {
	auth := f.auth.Authorize(ctx, req)
	return context.WithValue(ctx, AuthCtxKey, auth)
}

func secureWrap(actual endpoint.Endpoint, required bool) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		_, err := authenticated(ctx, required)
		if err != nil {
			return nil, err
		}

		return actual(ctx, req)
	}
}

func authenticated(ctx context.Context, strict bool) (*Authorization, error) {
	auth, ok := ctx.Value(AuthCtxKey).(Authorization)
	if strict && (!ok || auth.Error != nil) {
		return nil, AccessError{}
	}

	return &auth, nil
}

type AccessError struct {
	Err error
}

func (e AccessError) Error() string {
	if e.Err == nil {
		return "authentication-failed"
	}

	return e.Err.Error()
}

func (e AccessError) Code() int {
	return 401
}

const (
	AuthTokenHeaderName = "x-auth-token"
)

func ExtractApiToken(req *http.Request) string {
	return req.Header.Get(AuthTokenHeaderName)
}

func IsApiTokenCorrect(req *http.Request, shouldBe string) bool {
	return ExtractApiToken(req) == shouldBe
}
