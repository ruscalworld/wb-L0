package order

import (
	"errors"
	"fmt"
	"time"
)

type Order struct {
	OrderUID          string    `json:"order_uid" db:"order_uid"`
	TrackNumber       string    `json:"track_number" db:"track_number"`
	Entry             string    `json:"entry" db:"entry"`
	Delivery          *Delivery `json:"delivery" db:"delivery"`
	Payment           *Payment  `json:"payment" db:"payment"`
	Items             []*Item   `json:"items" db:"-"`
	Locale            string    `json:"locale" db:"locale"`
	InternalSignature string    `json:"internal_signature" db:"internal_signature"`
	CustomerID        string    `json:"customer_id" db:"customer_id"`
	DeliveryService   string    `json:"delivery_service" db:"delivery_service"`
	ShardKey          string    `json:"shardkey" db:"shardkey"`
	SmID              int64     `json:"sm_id" db:"sm_id"`
	DateCreated       time.Time `json:"date_created" db:"date_created"`
	OofShard          string    `json:"oof_shard" db:"oof_shard"`
}

func (o *Order) Validate() error {
	if o.OrderUID == "" {
		return errors.New("order uid is empty")
	}

	if o.Delivery == nil {
		return errors.New("delivery info is empty")
	}

	if o.Payment == nil {
		return errors.New("payment info is empty")
	}

	err := o.Payment.Validate()
	if err != nil {
		return fmt.Errorf("payment is invalid: %s", err)
	}

	if o.Items == nil {
		return errors.New("item list is nil")
	}

	for i, item := range o.Items {
		err := item.Validate()
		if err != nil {
			return fmt.Errorf("item %d is invalid: %s", i, err)
		}
	}

	return nil
}

type Delivery struct {
	ID      int64  `json:"-" db:"id"`
	Name    string `json:"name" db:"name"`
	Phone   string `json:"phone" db:"phone"`
	Zip     string `json:"zip" db:"zip"`
	City    string `json:"city" db:"city"`
	Address string `json:"address" db:"address"`
	Region  string `json:"region" db:"region"`
	Email   string `json:"email" db:"email"`
}

type Payment struct {
	Transaction  string  `json:"transaction" db:"transaction"`
	RequestID    string  `json:"request_id" db:"request_id"`
	Currency     string  `json:"currency" db:"currency"`
	Provider     string  `json:"provider" db:"provider"`
	Amount       float64 `json:"amount" db:"amount"`
	PaymentDt    int64   `json:"payment_dt" db:"payment_dt"`
	Bank         string  `json:"bank" db:"bank"`
	DeliveryCost float64 `json:"delivery_cost" db:"delivery_cost"`
	GoodsTotal   float64 `json:"goods_total" db:"goods_total"`
	CustomFee    float64 `json:"custom_fee" db:"custom_fee"`
}

func (p *Payment) Validate() error {
	if p.Transaction == "" {
		return errors.New("transaction is empty")
	}

	return nil
}

type Item struct {
	ChrtID      int64   `json:"chrt_id" db:"chrt_id"`
	TrackNumber string  `json:"track_number" db:"track_number"`
	Price       float64 `json:"price" db:"price"`
	RID         string  `json:"rid" db:"rid"`
	Name        string  `json:"name" db:"name"`
	Sale        float64 `json:"sale" db:"sale"`
	Size        string  `json:"size" db:"size"`
	TotalPrice  float64 `json:"total_price" db:"total_price"`
	NmID        int64   `json:"nm_id" db:"nm_id"`
	Brand       string  `json:"brand" db:"brand"`
	Status      int     `json:"status" db:"status"`
	OrderUID    string  `json:"-" db:"order_uid"`
}

func (i *Item) Validate() error {
	if i.ChrtID == 0 {
		return errors.New("illegal chrt_id (0)")
	}

	return nil
}
