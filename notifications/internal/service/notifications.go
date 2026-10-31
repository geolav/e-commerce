//// package service
////
//// import (
////	"bytes"
////	"context"
////	"encoding/json"
////	"fmt"
////	"net/http"
////	"strings"
////
////	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
////	"go.uber.org/zap"
////	"google.golang.org/protobuf/types/known/emptypb"
//// )
////
//// const failedStatusCode = 300
////
//// type notificationsServer struct {
////	notificationspb.UnimplementedNotificationsServer
////	callbackAddr string
////	logger       *zap.Logger
//// }
////
//// func NewNotificationsServer(callbackAddr string, logger *zap.Logger) *notificationsServer {
////	return &notificationsServer{
////		callbackAddr: callbackAddr,
////		logger:       logger,
////	}
//// }
////
//// type callbackPayload struct {
////	UserID  int64  `json:"user_id"`
////	OrderID int64  `json:"order_id"`
////	Status  string `json:"status"`
//// }
////
//// func (s *notificationsServer) Notify(_ context.Context, req *notificationspb.NotifyRequest) (*emptypb.Empty, error) {
////	callbackAddr := strings.TrimSpace(s.callbackAddr)
////	if callbackAddr == "" {
////		return &emptypb.Empty{}, nil
////	}
////
////	payload := callbackPayload{
////		UserID:  req.GetUserId(),
////		OrderID: req.GetOrderId(),
////		Status:  req.GetStatus(),
////	}
////
////	body, err := json.Marshal(payload)
////	if err != nil {
////		return nil, fmt.Errorf("failed to marshal payload: %w", err)
////	}
////
////	resp, err := http.Post("http://"+callbackAddr, "application/json", bytes.NewReader(body))
////	if err != nil {
////		return nil, fmt.Errorf("failed to send notification callback: %w", err)
////	}
////	defer resp.Body.Close()
////
////	if resp.StatusCode >= failedStatusCode {
////		return nil, fmt.Errorf("failed to send notification callback: callback returned status code %d", resp.StatusCode)
////	}
////
////	return &emptypb.Empty{}, nil
//// }
//
// package service
//
// import (
//	"bytes"
//	"context"
//	"encoding/json"
//	"fmt"
//	"net/http"
//	"strings"
//
//	"go.uber.org/zap"
// )
//
// const failedStatusCode = 300
//
// type NotificationPayload struct {
//	UserID  int64  `json:"user_id"`
//	OrderID int64  `json:"order_id"`
//	Status  string `json:"status"`
// }
//
// type Notifier struct {
//	callbackAddr string
//	logger       *zap.Logger
// }
//
// func NewNotifier(callbackAddr string, logger *zap.Logger) *Notifier {
//	return &Notifier{
//		callbackAddr: callbackAddr,
//		logger:       logger,
//	}
// }
//
// func (n *Notifier) Notify(_ context.Context, payload NotificationPayload) error {
//	callbackAddr := strings.TrimSpace(n.callbackAddr)
//	if callbackAddr == "" {
//		// Если всё ещё пусто, логируем и не отправляем
//		n.logger.Warn("callback address is empty, skipping notification")
//		return nil
//	}
//
//	// Формируем полный URL
//	url := "http://" + callbackAddr
//	n.logger.Info("sending notification", zap.String("url", url), zap.Any("payload", payload))
//
//	body, err := json.Marshal(payload)
//	if err != nil {
//		return fmt.Errorf("failed to marshal payload: %w", err)
//	}
//
//	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
//	if err != nil {
//		n.logger.Warn("failed to send notification", zap.Error(err))
//		return fmt.Errorf("failed to send notification callback: %w", err)
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode >= 300 {
//		n.logger.Warn("callback returned error status", zap.Int("status", resp.StatusCode))
//		return fmt.Errorf("callback returned status code %d", resp.StatusCode)
//	}
//
//	n.logger.Info("notification sent successfully", zap.Int("status", resp.StatusCode))
//	return nil
// }

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const failedStatusCode = 300

type NotificationPayload struct {
	UserID  int64  `json:"user_id"`
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

type Notifier struct {
	callbackAddr string
	logger       *zap.Logger
}

func NewNotifier(callbackAddr string, logger *zap.Logger) *Notifier {
	return &Notifier{
		callbackAddr: callbackAddr,
		logger:       logger,
	}
}

func (n *Notifier) Notify(ctx context.Context, payload NotificationPayload) error {
	callbackAddr := strings.TrimSpace(n.callbackAddr)

	// Если адрес не задан, используем адрес по умолчанию для тестов
	// if callbackAddr == "" {
	//	callbackAddr = "integration-tests:8080/callback"
	//	n.logger.Info("using default callback address", zap.String("addr", callbackAddr))
	// }

	if callbackAddr == "" {
		n.logger.Warn("callback address is empty, skipping notification")
		return nil
	}

	url := "http://" + callbackAddr
	n.logger.Info("sending notification",
		zap.String("url", url),
		zap.Int64("order_id", payload.OrderID),
		zap.Int64("user_id", payload.UserID),
		zap.String("status", payload.Status))

	body, err := json.Marshal(payload)
	if err != nil {
		n.logger.Error("failed to marshal payload", zap.Error(err))
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// resp, err := http.Post(url, "application/json", bytes.NewReader(body))

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		n.logger.Warn(
			"failed to send notification callback",
			zap.Error(err),
		)
		return fmt.Errorf("failed to send notification callback: %w", err)
	}

	if err != nil {
		n.logger.Warn("failed to send notification callback", zap.Error(err))
		return fmt.Errorf("failed to send notification callback: %w", err)
	}
	defer resp.Body.Close()

	n.logger.Info("callback response", zap.Int("status_code", resp.StatusCode))

	if resp.StatusCode >= failedStatusCode {
		return fmt.Errorf("callback returned status code %d", resp.StatusCode)
	}

	return nil
}
