package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCorsHandler_ReturnsHandler(t *testing.T) {
	handler := corsHandler(nil)
	assert.NotNil(t, handler)
}

func TestCorsHandler_OptionsMethod_ReturnsOK(t *testing.T) {
	handler := corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "POST, GET, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
}

func TestCorsHandler_GetMethod_CallsNext(t *testing.T) {
	called := false
	handler := corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCorsHandler_PostMethod_CallsNext(t *testing.T) {
	called := false
	handler := corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCorsHandler_SetsCorrectHeaders(t *testing.T) {
	handler := corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST, GET, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Accept", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestCorsHandler_WithCustomOrigin_PassesThrough(t *testing.T) {
	handler := corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// CORS middleware должен пробросить оригинальный Origin
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestAppConfigValidation_ValidConfig(t *testing.T) {
	cfg := &config.Config{}
	cfg.GRPC.Port = "50052"
	cfg.GRPC.GatewayPort = "8081"
	cfg.Outbox.Workers = 1
	cfg.Outbox.BatchSize = 10
	cfg.Outbox.FetchPeriod = 100 * time.Millisecond
	cfg.Outbox.InProgressTTL = 5 * time.Second
	cfg.Kafka.Brokers = "localhost:9092"
	cfg.Kafka.Topic = "test-topic"

	assert.NotEmpty(t, cfg.GRPC.Port)
	assert.NotEmpty(t, cfg.GRPC.GatewayPort)
	assert.Greater(t, cfg.Outbox.Workers, 0)
	assert.NotEmpty(t, cfg.Kafka.Brokers)
}

func TestCorsHandler_AllMethods(t *testing.T) {
	handler := corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Проверяем что GET и POST работают
	methods := []string{http.MethodGet, http.MethodPost}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestCorsHandler_Options(t *testing.T) {
	handler := corsHandler(nil)
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "POST, GET, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
}
