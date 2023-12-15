package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"wb-l0/internal/order"
)

// CachedRepository - репозиторий, который работает с двумя другими репозиториями. Один из них считается
// кэшем, другой - основной базой данных.
//
// При запросе информации сначала делается запрос в кэш, при отсутствии ошибок и наличии значения в кэше,
// это значение возвращается сразу же.
// При возникновении ошибок (в т.ч. если значение не найдено) производится запрос к основной базе данных,
// и значение возвращается оттуда, при этом оно помещается в кэш для ускорения работы последующих запросов.
// При отсутствии нужных данных в основной базе данных возвращается ошибка order.ErrNotFound.
type CachedRepository struct {
	database order.Repository
	cache    order.Repository
}

func NewCachedRepository(database order.Repository, cache order.Repository) *CachedRepository {
	return &CachedRepository{database: database, cache: cache}
}

func (c *CachedRepository) GetOrder(ctx context.Context, uid string) (*order.Order, error) {
	// Пробуем получить значение из кэша.
	o, err := c.cache.GetOrder(ctx, uid)
	if err == nil {
		return o, nil
	}

	// Проверяем, столкнулись мы с реальной ошибкой или же просто не смогли найти нужное значение.
	// Логируем ошибку, если она не связана с тем, что отсутствует значение в базе данных.
	if !errors.Is(err, order.ErrNotFound) {
		log.Printf("error fetching order %s from cache: %s\n", uid, err)
	}

	// Пробуем получить значение из основной базы данных.
	o, err = c.database.GetOrder(ctx, uid)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			return nil, order.ErrNotFound
		}

		return nil, fmt.Errorf("error fetching order %s from database: %s", uid, err)
	}

	// Сохраняем полученной из основной базы данных значение в кэш.
	_ = c.cache.CreateOrder(ctx, o)
	return o, nil
}

func (c *CachedRepository) CreateOrder(ctx context.Context, o *order.Order) error {
	err := c.database.CreateOrder(ctx, o)
	if err != nil {
		return fmt.Errorf("error saving order %s to database: %s", o.OrderUID, err)
	}

	err = c.cache.CreateOrder(ctx, o)
	if err != nil {
		log.Printf("error saving order %s to cache: %s\n", o.OrderUID, err)
	}

	return nil
}
