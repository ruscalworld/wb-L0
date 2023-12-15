package order

import "net/http"

type HttpHandlers interface {
	GetOrder(r *http.Request) (any, error)
}
