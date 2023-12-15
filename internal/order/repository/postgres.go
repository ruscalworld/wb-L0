package repository

import (
	"context"
	"errors"
	"fmt"

	"wb-l0/internal/config"
	"wb-l0/internal/order"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

// PostgresRepository - обёртка над соединением pgx, позволяющая сохранять в базе данных представленные в виде структур
// Go сущности, связанные с заказами, а также получать их. Использует четыре таблицы:
//
// orders непосредственно для хранения самих заказов;
// deliveries для хранения информации о доставках: orders.delivery_id -> deliveries.id;
// payments для хранения информации о платежах: orders.transaction -> payments.transaction;
// items для хранения самих товаров: items.order_uid -> orders.order_uid.
type PostgresRepository struct {
	conn *pgx.Conn
}

func NewPostgresRepository(conn *pgx.Conn) *PostgresRepository {
	return &PostgresRepository{conn: conn}
}

func NewPostgresRepositoryFromConfig(ctx context.Context, cfg config.PostgresConnection) (*PostgresRepository, error) {
	conn, err := pgx.Connect(ctx, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to postgres: %s", err)
	}
	return NewPostgresRepository(conn), nil
}

const getOrderQuery string = `
select o.order_uid,
       o.track_number,
       o.entry,
       o.locale,
       o.internal_signature,
       o.customer_id,
       o.delivery_service,
       o.shardkey,
       o.sm_id,
       o.date_created,
       o.oof_shard,
       d.name as "delivery.name",
       d.phone as "delivery.phone",
       d.zip as "delivery.zip",
       d.city as "delivery.city",
       d.address as "delivery.address",
       d.region as "delivery.region",
       d.email as "delivery.email",
       p.transaction as "payment.transaction",
       p.request_id as "payment.request_id",
       p.currency as "payment.currency",
       p.provider as "payment.provider",
       p.amount as "payment.amount",
       p.payment_dt as "payment.payment_dt",
       p.bank as "payment.bank",
       p.delivery_cost as "payment.delivery_cost",
       p.goods_total as "payment.goods_total",
       p.custom_fee as "payment.custom_fee"
from orders o
         join deliveries d on d.id = o.delivery_id
         join payments p on p.transaction = o.transaction
where o.order_uid = $1`

const getOrderItemsQuery = `
select chrt_id,
       track_number,
       price,
       rid,
       name,
       sale,
       size,
       total_price,
       nm_id,
       brand,
       status,
       order_uid
from items where order_uid = $1
`

func (r *PostgresRepository) GetOrder(ctx context.Context, uid string) (*order.Order, error) {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %s", err)
	}
	defer tx.Rollback(ctx)

	var o order.Order
	err = pgxscan.Get(ctx, tx, &o, getOrderQuery, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, order.ErrNotFound
		}

		return nil, fmt.Errorf("error fetching order from database: %s", err)
	}

	err = pgxscan.Select(ctx, tx, &o.Items, getOrderItemsQuery, uid)
	if err != nil {
		return nil, fmt.Errorf("error fetching order items from database: %s", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %s", err)
	}

	return &o, nil
}

const createOrderQuery = `
insert into orders
    (order_uid, track_number, entry, delivery_id, transaction, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

func (r *PostgresRepository) CreateOrder(ctx context.Context, o *order.Order) error {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %s", err)
	}
	defer tx.Rollback(ctx)

	err = r.createPayment(ctx, tx, o.Payment)
	if err != nil {
		return fmt.Errorf("error saving payment in database: %s", err)
	}

	err = r.createDelivery(ctx, tx, o.Delivery)
	if err != nil {
		return fmt.Errorf("error saving delivery in database: %s", err)
	}

	_, err = tx.Exec(
		ctx, createOrderQuery,
		o.OrderUID, o.TrackNumber, o.Entry, o.Delivery.ID, o.Payment.Transaction, o.Locale, o.InternalSignature,
		o.CustomerID, o.DeliveryService, o.ShardKey, o.SmID, o.DateCreated, o.OofShard,
	)
	if err != nil {
		return fmt.Errorf("error saving order in database: %s", err)
	}

	for _, item := range o.Items {
		item.OrderUID = o.OrderUID
		err := r.createItem(ctx, tx, item)
		if err != nil {
			return fmt.Errorf("error saving item %s in database: %s", item.ChrtID, err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error committing transaction: %s", err)
	}

	return nil
}

const createPaymentQuery = `
insert into payments
    (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

func (*PostgresRepository) createPayment(ctx context.Context, tx pgx.Tx, payment *order.Payment) error {
	_, err := tx.Exec(
		ctx, createPaymentQuery,
		payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDt,
		payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee,
	)

	return err
}

const createDeliveryQuery = `
insert into deliveries
	(name, phone, zip, city, address, region, email)
	values ($1, $2, $3, $4, $5, $6, $7) returning id`

func (*PostgresRepository) createDelivery(ctx context.Context, tx pgx.Tx, delivery *order.Delivery) error {
	err := tx.QueryRow(
		ctx, createDeliveryQuery,
		delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email,
	).Scan(&delivery.ID)

	return err
}

const createItemQuery = `
insert into items
    (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, order_uid)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

func (r *PostgresRepository) createItem(ctx context.Context, tx pgx.Tx, item *order.Item) error {
	_, err := tx.Exec(
		ctx, createItemQuery,
		item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice,
		item.NmID, item.Brand, item.Status, item.OrderUID,
	)

	return err
}
