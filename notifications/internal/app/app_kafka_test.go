package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/service"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUnmarshalNotificationPayload(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOrderID int64
		wantUserID  int64
		wantStatus  string
		wantErr     bool
	}{
		{
			name:        "valid awaiting_payment",
			input:       `{"user_id":900001,"order_id":42,"status":"awaiting_payment"}`,
			wantOrderID: 42,
			wantUserID:  900001,
			wantStatus:  "awaiting_payment",
		},
		{
			name:        "valid paid",
			input:       `{"user_id":1,"order_id":7,"status":"paid"}`,
			wantOrderID: 7,
			wantUserID:  1,
			wantStatus:  "paid",
		},
		{
			name:        "valid cancelled",
			input:       `{"user_id":2,"order_id":3,"status":"cancelled"}`,
			wantOrderID: 3,
			wantUserID:  2,
			wantStatus:  "cancelled",
		},
		{
			name:    "invalid json",
			input:   `not json`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload service.NotificationPayload
			err := json.Unmarshal([]byte(tt.input), &payload)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantOrderID, payload.OrderID)
				require.Equal(t, tt.wantUserID, payload.UserID)
				require.Equal(t, tt.wantStatus, payload.Status)
			}
		})
	}
}

func TestNotifierRetryLogic(t *testing.T) {
	tests := []struct {
		name        string
		cancelCtx   bool
		callbackURL string
		wantErr     bool
	}{
		{
			name:        "cancelled context returns error",
			cancelCtx:   true,
			callbackURL: "localhost:19999",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}

			notifier := service.NewNotifier(tt.callbackURL, zap.NewNop())
			err := notifier.Notify(ctx, service.NotificationPayload{
				OrderID: 1, UserID: 1, Status: "awaiting_payment",
			})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
