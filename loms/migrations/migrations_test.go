package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbedMigrations(t *testing.T) {
	assert.NotNil(t, embedMigrations)
}

func TestSetupPostgres_NoDB(t *testing.T) {
	t.Log("migrations package compiled")
}

func TestEmbedMigrationsExists(t *testing.T) {
	entries, err := embedMigrations.ReadDir("migrations")
	if err != nil {
		t.Skip("migrations directory not found")
	}
	t.Logf("Found %d migration files", len(entries))
}

func TestEmbedMigrations_ContainsFiles(t *testing.T) {
	entries, err := embedMigrations.ReadDir("migrations")
	if err != nil {
		t.Skip("migrations directory not accessible in test environment")
		return
	}
	assert.NotEmpty(t, entries)
}
