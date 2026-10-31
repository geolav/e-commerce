package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		wantGRPCPort string
		wantGWPort   string
		wantPGHost   string
	}{
		{
			name:         "default config loads correctly",
			wantGRPCPort: "50052",
			wantGWPort:   "8081",
			wantPGHost:   "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := New()
			require.NoError(t, err)
			require.NotNil(t, cfg)
			require.Equal(t, tt.wantGRPCPort, cfg.GRPC.Port)
			require.Equal(t, tt.wantGWPort, cfg.GRPC.GatewayPort)
			require.Equal(t, tt.wantPGHost, cfg.PG.Host)
		})
	}
}
