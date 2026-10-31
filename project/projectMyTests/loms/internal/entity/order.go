package entity

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Order struct {
	ID        int64
	Status    OrderStatus
	UserID    int64
	Items     []OrderItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrderItem struct {
	SKU   uint32
	Count uint32
}

type OutboxEvent struct {
	ID           int64
	Payload      map[string]interface{} // вместо полей order_id, user_id, status
	Attempts     int
	InProgressAt *time.Time
}

type OrderStatus string

const (
	OrderStatusNew         OrderStatus = "new"
	OrderStatusAwaitingPay OrderStatus = "awaiting_payment"
	OrderStatusPayed       OrderStatus = "paid"
	OrderStatusFailed      OrderStatus = "failed"
	OrderStatusCancelled   OrderStatus = "cancelled"
)

func (s *OrderStatus) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		*s = OrderStatus(v)
	case []byte:
		*s = OrderStatus(v)
	default:
		return fmt.Errorf("unsupported scan type for OrderStatus: %T", src)
	}
	return nil
}

func (s OrderStatus) Value() (driver.Value, error) {
	return string(s), nil
}
