package config

import (
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
)

func TestLoadSecretsFromGSM_Error(t *testing.T) {
	config := models.Config{
		MongoDB: models.MongoConfig{
			URI: "gsm:projects/project-id/secrets/secret-name/versions/latest",
		},
		Mnemonic: "gsm:projects/project-id/secrets/mnemonic-name/versions/latest",
	}

	configWithSecrets, err := loadSecretsFromGSM(config)

	assert.Equal(t, "gsm:projects/project-id/secrets/secret-name/versions/latest", configWithSecrets.MongoDB.URI)
	assert.Equal(t, "gsm:projects/project-id/secrets/mnemonic-name/versions/latest", configWithSecrets.Mnemonic)
	assert.Error(t, err)
}

func TestReadSecretFromGSM(t *testing.T) {
	client := &secretmanager.Client{}

	t.Run("Already set, skipping GSM read", func(t *testing.T) {
		value, err := readSecretFromGSM(client, "label", "value")
		assert.Equal(t, "value", value)
		assert.NoError(t, err)
	})

	// t.Run("Successfully read secret from GSM", func(t *testing.T) {
	// 	value := readSecretFromGSM(client, "label", "gsm:projects/project-id/secrets/secret-name/versions/latest")
	// 	assert.Equal(t, "gsm:projects/project-id/secrets/secret-name/versions/latest", value)
	// })
	//
	// t.Run("Failed to read GSM secret", func(t *testing.T) {
	// 	value := readSecretFromGSM(client, "label", "gsm:projects/project-id/secrets/invalid-secret-name/versions/latest")
	// 	assert.Equal(t, "", value)
	// })
}
