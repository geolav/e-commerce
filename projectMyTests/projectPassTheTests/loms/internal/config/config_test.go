package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cfg, err := New()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "50052", cfg.GRPC.Port)
	require.Equal(t, "8081", cfg.GRPC.GatewayPort)
	require.Equal(t, "localhost", cfg.PG.Host)
}
