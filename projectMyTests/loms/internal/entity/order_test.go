package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOrderStatus_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected OrderStatus
		wantErr  bool
	}{
		{"string paid", "paid", OrderStatusPayed, false},
		{"string cancelled", "cancelled", OrderStatusCancelled, false},
		{"string awaiting_payment", "awaiting_payment", OrderStatusAwaitingPay, false},
		{"bytes cancelled", []byte("cancelled"), OrderStatusCancelled, false},
		{"invalid type", 123, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s OrderStatus
			err := s.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, s)
			}
		})
	}
}

func TestOrderStatus_Value(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected string
	}{
		{OrderStatusNew, "new"},
		{OrderStatusAwaitingPay, "awaiting_payment"},
		{OrderStatusPayed, "paid"},
		{OrderStatusFailed, "failed"},
		{OrderStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			v, err := tt.status.Value()
			require.NoError(t, err)
			require.Equal(t, tt.expected, v)
		})
	}
}

func TestOrderStruct(t *testing.T) {
	tests := []struct {
		name      string
		order     Order
		wantID    int64
		wantItems int
	}{
		{
			name: "basic order",
			order: Order{
				ID:        1,
				Status:    OrderStatusNew,
				UserID:    42,
				Items:     []OrderItem{{SKU: 1, Count: 2}},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantID:    1,
			wantItems: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantID, tt.order.ID)
			require.Len(t, tt.order.Items, tt.wantItems)
		})
	}
}
