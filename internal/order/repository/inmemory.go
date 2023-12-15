package repository

import (
	"context"
	"errors"
	"sync"

	"wb-l0/internal/order"
)

// InMemoryRepository - репозиторий, хранящий все данные в памяти, используя sync.Map.
type InMemoryRepository struct {
	store *sync.Map
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		store: &sync.Map{},
	}
}

func (i *InMemoryRepository) GetOrder(_ context.Context, uid string) (*order.Order, error) {
	value, ok := i.store.Load(uid)
	if !ok {
		return nil, order.ErrNotFound
	}

	return value.(*order.Order), nil
}

func (i *InMemoryRepository) CreateOrder(_ context.Context, o *order.Order) error {
	if _, ok := i.store.Load(o.OrderUID); ok {
		return errors.New("order with the same id already exists")
	}

	i.store.Store(o.OrderUID, o)
	return nil
}
