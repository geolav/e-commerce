package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
)

func TestNewNotificationsServer(t *testing.T) {
	logger := zap.NewNop()
	srv := NewNotificationsServer("localhost:8080", logger)
	require.NotNil(t, srv)
}

func TestNotify_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// убираем http:// из адреса
	addr := server.URL[len("http://"):]

	svc := NewNotificationsServer(addr, zap.NewNop())

	_, err := svc.Notify(context.Background(), &notificationspb.NotifyRequest{
		UserId:  1,
		OrderId: 42,
		Status:  "awaiting_payment",
	})
	require.NoError(t, err)
}

func TestNotify_EmptyCallbackAddr(t *testing.T) {
	svc := NewNotificationsServer("", zap.NewNop())

	_, err := svc.Notify(context.Background(), &notificationspb.NotifyRequest{
		UserId:  1,
		OrderId: 42,
		Status:  "awaiting_payment",
	})
	require.NoError(t, err)
}

func TestNotify_CallbackFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	addr := server.URL[len("http://"):]
	svc := NewNotificationsServer(addr, zap.NewNop())

	_, err := svc.Notify(context.Background(), &notificationspb.NotifyRequest{
		UserId:  1,
		OrderId: 42,
		Status:  "awaiting_payment",
	})
	require.Error(t, err)
}

func TestNotify_ServerUnreachable(t *testing.T) {
	svc := NewNotificationsServer("localhost:19999", zap.NewNop())

	_, err := svc.Notify(context.Background(), &notificationspb.NotifyRequest{
		UserId:  1,
		OrderId: 42,
		Status:  "awaiting_payment",
	})
	require.Error(t, err)
}
