package config

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
)

type mockSecretManagerClient struct {
	mockStore map[string]string
	mockError map[string]error
}

func (m *mockSecretManagerClient) GetSecretValue(secretName string) string {
	if m.mockStore == nil {
		m.mockStore = make(map[string]string)
	}
	return m.mockStore[secretName]
}

func (m *mockSecretManagerClient) SetSecretValue(secretName string, value string) {
	if m.mockStore == nil {
		m.mockStore = make(map[string]string)
	}
	m.mockStore[secretName] = value
}

func (m *mockSecretManagerClient) GetError(secretName string) error {
	if m.mockError == nil {
		m.mockError = make(map[string]error)
	}
	return m.mockError[secretName]
}

func (m *mockSecretManagerClient) SetError(secretName string, err error) {
	if m.mockError == nil {
		m.mockError = make(map[string]error)
	}
	m.mockError[secretName] = err
}

func (m *mockSecretManagerClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if err := m.GetError(req.Name); err != nil {
		return nil, err
	}
	return &secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(m.GetSecretValue(req.Name)),
		},
	}, nil
}

func (m *mockSecretManagerClient) Close() error {
	return nil
}

func TestNewSecretManagerClientShouldErrorWithoutCreds(t *testing.T) {
	client, err := NewSecretManagerClient()
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestLoadSecretsFromGSM_Error(t *testing.T) {
	t.Run("Not GSM secrets", func(t *testing.T) {
		config := models.Config{
			MongoDB: models.MongoConfig{
				URI: "mongodb://localhost:27017",
			},
			Mnemonic: "mnemonic",
		}

		configWithSecrets, err := loadSecretsFromGSM(config)

		assert.Equal(t, "mongodb://localhost:27017", configWithSecrets.MongoDB.URI)
		assert.Equal(t, "mnemonic", configWithSecrets.Mnemonic)
		assert.NoError(t, err)
	})

	t.Run("Failed to create GSM client", func(t *testing.T) {
		oldNewSecretManagerClient := NewSecretManagerClient
		NewSecretManagerClient = func() (SecretManagerClient, error) {
			return nil, errors.New("error")
		}
		defer func() { NewSecretManagerClient = oldNewSecretManagerClient }()
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
	})

	t.Run("Failed to read MongoDB URI", func(t *testing.T) {
		oldNewSecretManagerClient := NewSecretManagerClient
		NewSecretManagerClient = func() (SecretManagerClient, error) {
			client := &mockSecretManagerClient{}
			client.SetError("projects/project-id/secrets/secret-name/versions/latest", errors.New("error"))
			return client, nil
		}
		defer func() { NewSecretManagerClient = oldNewSecretManagerClient }()

		config := models.Config{
			MongoDB: models.MongoConfig{
				URI: "gsm:projects/project-id/secrets/secret-name/versions/latest",
			},
			Mnemonic: "gsm:projects/project-id/secrets/mnemonic-name/versions/latest",
		}

		configWithSecrets, err := loadSecretsFromGSM(config)

		assert.Equal(t, "", configWithSecrets.MongoDB.URI)
		assert.Equal(t, "gsm:projects/project-id/secrets/mnemonic-name/versions/latest", configWithSecrets.Mnemonic)
		assert.Error(t, err)

	})

	t.Run("Failed to read mnemonic", func(t *testing.T) {
		oldNewSecretManagerClient := NewSecretManagerClient
		NewSecretManagerClient = func() (SecretManagerClient, error) {
			client := &mockSecretManagerClient{}
			client.SetSecretValue("projects/project-id/secrets/secret-name/versions/latest", "mongodb://localhost:27017")
			client.SetError("projects/project-id/secrets/mnemonic-name/versions/latest", errors.New("error"))
			return client, nil
		}
		defer func() { NewSecretManagerClient = oldNewSecretManagerClient }()

		config := models.Config{
			MongoDB: models.MongoConfig{
				URI: "gsm:projects/project-id/secrets/secret-name/versions/latest",
			},
			Mnemonic: "gsm:projects/project-id/secrets/mnemonic-name/versions/latest",
		}

		configWithSecrets, err := loadSecretsFromGSM(config)

		assert.Equal(t, "mongodb://localhost:27017", configWithSecrets.MongoDB.URI)
		assert.Equal(t, "", configWithSecrets.Mnemonic)
		assert.Error(t, err)
	})

	t.Run("Successfully read all secrets", func(t *testing.T) {
		oldNewSecretManagerClient := NewSecretManagerClient
		NewSecretManagerClient = func() (SecretManagerClient, error) {
			client := &mockSecretManagerClient{}
			client.SetSecretValue("projects/project-id/secrets/secret-name/versions/latest", "mongodb://localhost:27017")
			client.SetSecretValue("projects/project-id/secrets/mnemonic-name/versions/latest", "mnemonic")
			return client, nil
		}
		defer func() { NewSecretManagerClient = oldNewSecretManagerClient }()

		config := models.Config{
			MongoDB: models.MongoConfig{
				URI: "gsm:projects/project-id/secrets/secret-name/versions/latest",
			},
			Mnemonic: "gsm:projects/project-id/secrets/mnemonic-name/versions/latest",
		}

		configWithSecrets, err := loadSecretsFromGSM(config)

		assert.Equal(t, "mongodb://localhost:27017", configWithSecrets.MongoDB.URI)
		assert.Equal(t, "mnemonic", configWithSecrets.Mnemonic)
		assert.NoError(t, err)
	})
}

func TestReadSecretFromGSM(t *testing.T) {
	client := &mockSecretManagerClient{}

	t.Run("Already set, skipping GSM read", func(t *testing.T) {
		value, err := readSecretFromGSM(client, "label", "value")
		assert.Equal(t, "value", value)
		assert.NoError(t, err)
	})

	t.Run("Successfully read secret from GSM", func(t *testing.T) {
		secretName := "gsm:projects/project-id/secrets/secret-name/versions/latest"
		client.SetSecretValue(secretName[4:], "secret-value")
		value, err := readSecretFromGSM(client, "label", secretName)
		assert.Equal(t, "secret-value", value)
		assert.NoError(t, err)
	})

	t.Run("Failed to read GSM secret", func(t *testing.T) {
		secretName := "gsm:projects/project-id/secrets/secret-name/versions/latest"
		client.SetError(secretName[4:], errors.New("error"))
		value, err := readSecretFromGSM(client, "label", secretName)
		assert.Equal(t, "", value)
		assert.Error(t, err)
	})
}
