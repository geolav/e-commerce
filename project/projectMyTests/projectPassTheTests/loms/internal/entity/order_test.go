package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOrderStatus_Scan_String(t *testing.T) {
	var s OrderStatus
	err := s.Scan("paid")
	require.NoError(t, err)
	require.Equal(t, OrderStatusPayed, s)
}

func TestOrderStatus_Scan_Bytes(t *testing.T) {
	var s OrderStatus
	err := s.Scan([]byte("cancelled"))
	require.NoError(t, err)
	require.Equal(t, OrderStatusCancelled, s)
}

func TestOrderStatus_Scan_Invalid(t *testing.T) {
	var s OrderStatus
	err := s.Scan(123)
	require.Error(t, err)
}

func TestOrderStatus_Value(t *testing.T) {
	s := OrderStatusAwaitingPay
	v, err := s.Value()
	require.NoError(t, err)
	require.Equal(t, "awaiting_payment", v)
}

func TestOrderStatus_AllConstants(t *testing.T) {
	tests := []OrderStatus{
		OrderStatusNew,
		OrderStatusAwaitingPay,
		OrderStatusPayed,
		OrderStatusFailed,
		OrderStatusCancelled,
	}

	for _, s := range tests {
		v, err := s.Value()
		require.NoError(t, err)
		require.Equal(t, string(s), v)
	}
}

func TestOrderStructs(t *testing.T) {
	order := Order{
		ID:     1,
		Status: OrderStatusNew,
		UserID: 42,
		Items: []OrderItem{
			{SKU: 1, Count: 2},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.Equal(t, int64(1), order.ID)
	require.Equal(t, OrderStatusNew, order.Status)
	require.Equal(t, int64(42), order.UserID)
	require.Len(t, order.Items, 1)
}
