package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

func TestInitConfig(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	t.Run("Invalid yaml file provided", func(t *testing.T) {
		assert.Panics(t, func() {
			InitConfig("random.yaml", "random.env")
		})
	})
	t.Run("Yaml file provided, but invalid env", func(t *testing.T) {
		yamlContent := `
mongodb:
  uri: "gsm://localhost:27017"
  database: "test"
  timeout_ms: 1000
`
		err := os.WriteFile("test.yaml", []byte(yamlContent), 0644)
		assert.NoError(t, err)
		defer os.Remove("test.yaml")

		assert.Panics(t, func() {
			InitConfig("test.yaml", "random.env")
		})
	})

	t.Run("Valid files but not gsm creds", func(t *testing.T) {
		yamlContent := `
mongodb:
  uri: "gsm://localhost:27017"
  database: "test"
  timeout_ms: 1000
`
		err := os.WriteFile("test.yaml", []byte(yamlContent), 0644)
		assert.NoError(t, err)
		defer os.Remove("test.yaml")

		assert.Panics(t, func() {
			InitConfig("test.yaml", "")
		})
	})

	t.Run("Valid files but invalid config", func(t *testing.T) {
		yamlContent := `
mongodb:
  uri: "mongodb://localhost:27017"
  database: "test"
  timeout_ms: 1000
`
		err := os.WriteFile("test.yaml", []byte(yamlContent), 0644)
		assert.NoError(t, err)
		defer os.Remove("test.yaml")

		envContent := `
		NUM_ETHEREUM_NETWORKS=2
		ETHEREUM_NETWORKS_0_CONFIRMATIONS=1
		ETHEREUM_NETWORKS_1_CONFIRMATIONS=2
		COSMOS_NETWORK_CONFIRMATIONS=3
		`
		err = os.WriteFile("test.env", []byte(envContent), 0644)
		assert.NoError(t, err)
		defer os.Remove("test.env")

		assert.Panics(t, func() {
			InitConfig("test.yaml", "test.env")
		})

		os.Unsetenv("NUM_ETHEREUM_NETWORKS")
		os.Unsetenv("ETHEREUM_NETWORKS_0_CONFIRMATIONS")
		os.Unsetenv("ETHEREUM_NETWORKS_1_CONFIRMATIONS")
		os.Unsetenv("COSMOS_NETWORK_CONFIRMATIONS")
	})

	t.Run("Valid files valid config", func(t *testing.T) {

		config := validConfig()
		assert.NotNil(t, config)

		yamlContent, err := yaml.Marshal(config)
		assert.NoError(t, err)

		err = os.WriteFile("test.yaml", []byte(yamlContent), 0644)
		assert.NoError(t, err)
		defer os.Remove("test.yaml")

		validatedConfig := InitConfig("test.yaml", "")
		assert.NotNil(t, validatedConfig)
		assert.Equal(t, config, validatedConfig)
	})

}
