package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNotify(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func() (addr string, cleanup func())
		wantErr     bool
	}{
		{
			name: "success 200",
			setupServer: func() (string, func()) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				return srv.URL[len("http://"):], srv.Close
			},
		},
		{
			name: "callback returns 503",
			setupServer: func() (string, func()) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusServiceUnavailable)
				}))
				return srv.URL[len("http://"):], srv.Close
			},
			wantErr: true,
		},
		{
			name: "empty addr skips",
			setupServer: func() (string, func()) {
				return "", func() {}
			},
			wantErr: false,
		},
		{
			name: "unreachable server",
			setupServer: func() (string, func()) {
				return "localhost:19999", func() {}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, cleanup := tt.setupServer()
			defer cleanup()

			notifier := NewNotifier(addr, zap.NewNop())
			err := notifier.Notify(context.Background(), NotificationPayload{
				UserID: 1, OrderID: 42, Status: "awaiting_payment",
			})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
