package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cfg, err := New()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "50053", cfg.GRPC.Port)
}
