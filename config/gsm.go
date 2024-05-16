package config

import (
	"context"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

func isGSMValue(value string) bool {
	return len(value) > 4 &&
		value[0:4] == "gsm:"
}

// if env variable is gsm:secret-name, read the secret from Google Secret Manager
func readSecretFromGSM(client *secretmanager.Client, label string, value string) string {
	if !isGSMValue(value) {
		log.Debugf("[CONFIG] Not reading %s from GSM", label)
		return value
	}
	name := value[4:]
	log.Debugf("[CONFIG] Reading %s from GSM for %s", name, label)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(context.Background(), req)
	if err != nil {
		log.Errorf("[CONFIG] Failed to access secret %s: %v", name, err)
		return ""
	}

	log.Debugf("[CONFIG] Successfully read %s from GSM for %s", name, label)
	return string(result.Payload.Data)
}

func loadSecretsFromGSM(config models.Config) models.Config {
	log.Debugf("[CONFIG] Loading secrets from GSM")
	configWithSecrets := config

	if !isGSMValue(config.MongoDB.URI) && !isGSMValue(config.Mnemonic) {
		log.Debugf("[CONFIG] No secrets to load from GSM")
		return configWithSecrets
	}

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Errorf("[CONFIG] Failed to create secretmanager client: %v", err)
		return configWithSecrets
	}
	defer client.Close()

	configWithSecrets.MongoDB.URI = readSecretFromGSM(client, "MongoDB.URI", config.MongoDB.URI)

	configWithSecrets.Mnemonic = readSecretFromGSM(client, "Mnemonic", config.Mnemonic)

	log.Debugf("[CONFIG] Successfully loaded secrets from GSM")
	return configWithSecrets
}
