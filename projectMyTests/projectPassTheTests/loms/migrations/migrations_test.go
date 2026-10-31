package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbedMigrations(t *testing.T) {
	files, err := embedMigrations.ReadDir("migrations")
	require.NoError(t, err)
	require.NotEmpty(t, files)
}
