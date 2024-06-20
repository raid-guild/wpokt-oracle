package config

import (
	"os"
	"testing"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Run("No env file provided", func(t *testing.T) {
		config, err := loadConfigFromEnv("")
		assert.Equal(t, models.Config{}, config)
		assert.NoError(t, err)
	})

	t.Run("Env file provided", func(t *testing.T) {
		envContent := `
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=testdb
`
		err := os.WriteFile(".test.env", []byte(envContent), 0644)
		assert.NoError(t, err)
		defer os.Remove(".test.env")

		config, err := loadConfigFromEnv(".test.env")
		assert.NoError(t, err)
		assert.Equal(t, "mongodb://localhost:27017", config.MongoDB.URI)
		assert.Equal(t, "testdb", config.MongoDB.Database)
	})

	t.Run("Error loading env file", func(t *testing.T) {
		config, err := loadConfigFromEnv("nonexistent.test.env")
		assert.Equal(t, models.Config{}, config)
		assert.Error(t, err)
	})

	t.Run("Valid env file with invalid values", func(t *testing.T) {
		envContent := `
MONGODB_TIMEOUT_MS=invalid
`
		err := os.WriteFile(".test.env", []byte(envContent), 0644)
		assert.NoError(t, err)
		defer os.Remove(".test.env")

		config, err := loadConfigFromEnv(".test.env")
		assert.NoError(t, err)
		assert.Equal(t, uint64(0), config.MongoDB.TimeoutMS)
	})
}
