package converter

import (
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"github.com/stretchr/testify/require"
)

func TestFromOrderStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    entity.OrderStatus
		expected lomspb.OrderStatus
	}{
		{"new", entity.OrderStatusNew, lomspb.OrderStatus_ORDER_STATUS_NEW},
		{"awaiting_payment", entity.OrderStatusAwaitingPay, lomspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT},
		{"failed", entity.OrderStatusFailed, lomspb.OrderStatus_ORDER_STATUS_FAILED},
		{"paid", entity.OrderStatusPayed, lomspb.OrderStatus_ORDER_STATUS_PAID},
		{"cancelled", entity.OrderStatusCancelled, lomspb.OrderStatus_ORDER_STATUS_CANCELLED},
		{"unknown defaults to unspecified", entity.OrderStatus("unknown"), lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, FromOrderStatus(tt.input))
		})
	}
}

func TestToOrderStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    lomspb.OrderStatus
		expected entity.OrderStatus
	}{
		{"new", lomspb.OrderStatus_ORDER_STATUS_NEW, entity.OrderStatusNew},
		{"awaiting_payment", lomspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT, entity.OrderStatusAwaitingPay},
		{"failed", lomspb.OrderStatus_ORDER_STATUS_FAILED, entity.OrderStatusFailed},
		{"paid", lomspb.OrderStatus_ORDER_STATUS_PAID, entity.OrderStatusPayed},
		{"cancelled", lomspb.OrderStatus_ORDER_STATUS_CANCELLED, entity.OrderStatusCancelled},
		{"unspecified defaults to new", lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED, entity.OrderStatusNew},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, ToOrderStatus(tt.input))
		})
	}
}
