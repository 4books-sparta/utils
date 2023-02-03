package --service-name--


type Middleware func(Service) Service

type noopMiddleware struct {
	next Service
}

