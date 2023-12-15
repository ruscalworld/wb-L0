package http

import (
	"net/http"
	"wb-l0/pkg/httperrors"

	"wb-l0/internal/order"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	orderRepository order.Repository
}

func NewOrderHandler(orderRepository order.Repository) *OrderHandler {
	return &OrderHandler{orderRepository: orderRepository}
}

func (h *OrderHandler) GetOrder(r *http.Request) (any, error) {
	orderId := chi.URLParam(r, "id")
	o, err := h.orderRepository.GetOrder(r.Context(), orderId)
	if err != nil {
		if err == order.ErrNotFound {
			return nil, httperrors.ErrNotFound
		}

		return nil, err
	}

	return o, nil
}
