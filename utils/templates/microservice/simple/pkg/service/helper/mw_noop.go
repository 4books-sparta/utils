package helper

type Middleware func(Service) Service

type noopMiddleware struct {
	next Service
}
