package converter

import (
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	pbloms "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
)

func FromOrderStatus(status entity.OrderStatus) pbloms.OrderStatus {
	switch status {
	case entity.OrderStatusNew:
		return pbloms.OrderStatus_ORDER_STATUS_NEW
	case entity.OrderStatusAwaitingPay:
		return pbloms.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT
	case entity.OrderStatusFailed:
		return pbloms.OrderStatus_ORDER_STATUS_FAILED
	case entity.OrderStatusPayed:
		return pbloms.OrderStatus_ORDER_STATUS_PAID
	case entity.OrderStatusCancelled:
		return pbloms.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return pbloms.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func ToOrderStatus(status pbloms.OrderStatus) entity.OrderStatus {
	switch status {
	case pbloms.OrderStatus_ORDER_STATUS_NEW:
		return entity.OrderStatusNew
	case pbloms.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT:
		return entity.OrderStatusAwaitingPay
	case pbloms.OrderStatus_ORDER_STATUS_FAILED:
		return entity.OrderStatusFailed
	case pbloms.OrderStatus_ORDER_STATUS_PAID:
		return entity.OrderStatusPayed
	case pbloms.OrderStatus_ORDER_STATUS_CANCELLED:
		return entity.OrderStatusCancelled
	default:
		return entity.OrderStatusNew
	}
}
