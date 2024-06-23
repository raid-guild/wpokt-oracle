package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

func TestGetUint64Env(t *testing.T) {
	t.Run("Valid env variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "123")
		defer os.Unsetenv("TEST_VAR")

		val := getUint64Env("TEST_VAR")
		assert.Equal(t, uint64(123), val)
	})

	t.Run("Invalid env variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "invalid")
		defer os.Unsetenv("TEST_VAR")

		val := getUint64Env("TEST_VAR")
		assert.Equal(t, uint64(0), val)
	})
}

func TestGetBoolEnv(t *testing.T) {
	t.Run("Valid env variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "true")
		defer os.Unsetenv("TEST_VAR")

		val := getBoolEnv("TEST_VAR")
		assert.True(t, val)
	})

	t.Run("Invalid env variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "invalid")
		defer os.Unsetenv("TEST_VAR")

		val := getBoolEnv("TEST_VAR")
		assert.False(t, val)
	})
}

func TestStringEnv(t *testing.T) {
	t.Run("Valid env variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test")
		defer os.Unsetenv("TEST_VAR")

		val := getStringEnv("TEST_VAR")
		assert.Equal(t, "test", val)
	})

	t.Run("Invalid env variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "")
		defer os.Unsetenv("TEST_VAR")

		val := getStringEnv("TEST_VAR")
		assert.Equal(t, "", val)
	})
}

func TestStringArrayEnv(t *testing.T) {
	t.Run("Valid env variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test1,test2")
		defer os.Unsetenv("TEST_VAR")

		val := getStringArrayEnv("TEST_VAR")
		assert.Equal(t, []string{"test1", "test2"}, val)
	})

	t.Run("Invalid env variable", func(t *testing.T) {
		os.Setenv("TEST_VAR", "")
		defer os.Unsetenv("TEST_VAR")

		val := getStringArrayEnv("TEST_VAR")
		var expectedVal []string = nil
		assert.Equal(t, expectedVal, val)
	})
}

func TestGetArrayLengthEnv(t *testing.T) {
	t.Run("Valid env variable", func(t *testing.T) {
		os.Setenv("TEST_VAR_ONE", "test1")
		os.Setenv("TEST_VAR_TWO", "test2")

		defer os.Unsetenv("TEST_VAR_ONE")
		defer os.Unsetenv("TEST_VAR_TWO")

		val := getArrayLengthEnv("TEST_VAR")
		assert.Equal(t, 2, val)
	})

	t.Run("Invalid env variable", func(t *testing.T) {
		val := getArrayLengthEnv("TEST_VAR")
		assert.Equal(t, 0, val)
	})
}

func TestLoadConfigFromEnv(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

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

	t.Run("Valid env file with invalid mongodb timeout", func(t *testing.T) {
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

	t.Run("Valid env file with invalid num eth networks", func(t *testing.T) {
		envContent := `
NUM_ETHEREUM_NETWORKS=2
`
		err := os.WriteFile(".test.env", []byte(envContent), 0644)
		assert.NoError(t, err)
		defer os.Remove(".test.env")

		assert.Panics(t, func() {
			loadConfigFromEnv(".test.env")
		})

		os.Unsetenv("NUM_ETHEREUM_NETWORKS")
	})

	t.Run("Valid env file with valid eth networks", func(t *testing.T) {
		envContent := `
NUM_ETHEREUM_NETWORKS=2
ETHEREUM_NETWORKS_0_CONFIRMATIONS=1
ETHEREUM_NETWORKS_1_CONFIRMATIONS=2
COSMOS_NETWORK_CONFIRMATIONS=3
`
		err := os.WriteFile(".test.env", []byte(envContent), 0644)
		assert.NoError(t, err)
		defer os.Remove(".test.env")

		config, err := loadConfigFromEnv(".test.env")

		assert.NoError(t, err)
		assert.Equal(t, 2, len(config.EthereumNetworks))
		assert.Equal(t, uint64(1), config.EthereumNetworks[0].Confirmations)
		assert.Equal(t, uint64(2), config.EthereumNetworks[1].Confirmations)
		assert.Equal(t, uint64(3), config.CosmosNetwork.Confirmations)

		os.Unsetenv("NUM_ETHEREUM_NETWORKS")
		os.Unsetenv("ETHEREUM_NETWORKS_0_CONFIRMATIONS")
		os.Unsetenv("ETHEREUM_NETWORKS_1_CONFIRMATIONS")
		os.Unsetenv("COSMOS_NETWORK_CONFIRMATIONS")
	})
}
