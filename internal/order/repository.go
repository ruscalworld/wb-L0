package order

import (
	"context"
	"net/http"

	"wb-l0/pkg/httperrors"
)

var (
	ErrNotFound = httperrors.NewHttpError("order with provided UID was not found in repository", http.StatusNotFound)
)

type Repository interface {
	GetOrder(ctx context.Context, uid string) (*Order, error)
	CreateOrder(ctx context.Context, order *Order) error
}
