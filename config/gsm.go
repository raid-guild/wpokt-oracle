package config

import (
	"context"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	"github.com/dan13ram/wpokt-oracle/models"
)

func isGSMValue(value string) bool {
	return len(value) > 4 &&
		value[0:4] == "gsm:"
}

// if env variable is gsm:secret-name, read the secret from Google Secret Manager
func readSecretFromGSM(client *secretmanager.Client, label string, value string) (string, error) {
	if !isGSMValue(value) {
		logger.
			WithField("config", label).
			Debugf("Already set, skipping GSM read")
		return value, nil
	}
	name := value[4:]
	logger.
		WithField("config", label).
		WithField("secret", name).
		Debugf("reading GSM secret")
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(context.Background(), req)
	if err != nil {
		logger.
			WithField("config", label).
			WithField("secret", name).
			WithError(err).
			Errorf("Failed to read GSM secret")
		return "", err
	}

	logger.
		WithField("config", label).
		WithField("secret", name).
		Debugf("Successfully read secret from GSM")
	return string(result.Payload.Data), nil
}

func loadSecretsFromGSM(config models.Config) (models.Config, error) {
	logger.Debugf("Loading secrets from GSM")
	configWithSecrets := config

	if !isGSMValue(config.MongoDB.URI) && !isGSMValue(config.Mnemonic) {
		logger.Debugf("No secrets to load from GSM")
		return configWithSecrets, nil
	}

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		logger.
			WithError(err).
			Errorf("Failed to create secretmanager client")
		return configWithSecrets, err
	}
	defer client.Close()

	configWithSecrets.MongoDB.URI, err = readSecretFromGSM(client, "MongoDB.URI", config.MongoDB.URI)
	if err != nil {
		return configWithSecrets, err
	}

	configWithSecrets.Mnemonic, err = readSecretFromGSM(client, "Mnemonic", config.Mnemonic)
	if err != nil {
		return configWithSecrets, err
	}

	logger.Debugf("Successfully loaded secrets from GSM")
	return configWithSecrets, nil
}
