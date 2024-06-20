package config

import (
	"os"
	"testing"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFromYamlFile(t *testing.T) {
	t.Run("No yaml file provided", func(t *testing.T) {
		config, err := loadConfigFromYamlFile("")
		assert.Equal(t, models.Config{}, config)
		assert.NoError(t, err)
	})
	t.Run("Yaml file provided", func(t *testing.T) {
		yamlContent := `
mongodb:
  uri: "mongodb://localhost:27017"
  database: "test"
  timeout_ms: 1000
`
		err := os.WriteFile("test.yaml", []byte(yamlContent), 0644)
		assert.NoError(t, err)
		defer os.Remove("test.yaml")

		config, err := loadConfigFromYamlFile("test.yaml")
		assert.NoError(t, err)
		assert.Equal(t, "mongodb://localhost:27017", config.MongoDB.URI)
		assert.Equal(t, "test", config.MongoDB.Database)
		assert.Equal(t, uint64(1000), config.MongoDB.TimeoutMS)
	})

	t.Run("Error reading yaml file", func(t *testing.T) {
		config, err := loadConfigFromYamlFile("nonexistent.yaml")
		assert.Error(t, err)
		assert.Equal(t, models.Config{}, config)
	})

	t.Run("Error unmarshalling yaml file", func(t *testing.T) {
		err := os.WriteFile("invalid.yaml", []byte("invalid content"), 0644)
		assert.NoError(t, err)
		defer os.Remove("invalid.yaml")

		config, err := loadConfigFromYamlFile("invalid.yaml")
		assert.Error(t, err)
		assert.Equal(t, models.Config{}, config)
	})
}
