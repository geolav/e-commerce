package converter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
)

func TestFromOrderStatus(t *testing.T) {
	cases := []struct {
		input    entity.OrderStatus
		expected lomspb.OrderStatus
	}{
		{entity.OrderStatusNew, lomspb.OrderStatus_ORDER_STATUS_NEW},
		{entity.OrderStatusAwaitingPay, lomspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT},
		{entity.OrderStatusFailed, lomspb.OrderStatus_ORDER_STATUS_FAILED},
		{entity.OrderStatusPayed, lomspb.OrderStatus_ORDER_STATUS_PAID},
		{entity.OrderStatusCancelled, lomspb.OrderStatus_ORDER_STATUS_CANCELLED},
	}

	for _, c := range cases {
		result := FromOrderStatus(c.input)
		require.Equal(t, c.expected, result)
	}
}

func TestToOrderStatus(t *testing.T) {
	cases := []struct {
		input    lomspb.OrderStatus
		expected entity.OrderStatus
	}{
		{lomspb.OrderStatus_ORDER_STATUS_NEW, entity.OrderStatusNew},
		{lomspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT, entity.OrderStatusAwaitingPay},
		{lomspb.OrderStatus_ORDER_STATUS_FAILED, entity.OrderStatusFailed},
		{lomspb.OrderStatus_ORDER_STATUS_PAID, entity.OrderStatusPayed},
		{lomspb.OrderStatus_ORDER_STATUS_CANCELLED, entity.OrderStatusCancelled},
	}

	for _, c := range cases {
		result := ToOrderStatus(c.input)
		require.Equal(t, c.expected, result)
	}
}

func TestFromOrderStatus_Default(t *testing.T) {
	result := FromOrderStatus(entity.OrderStatus("unknown"))
	require.Equal(t, lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED, result)
}

func TestToOrderStatus_Default(t *testing.T) {
	result := ToOrderStatus(lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED)
	require.Equal(t, entity.OrderStatusNew, result)
}
