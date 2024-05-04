package config

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

// if env variable is gsm:secret-name, read the secret from Google Secret Manager
func readSecretFromGSM(client *secretmanager.Client, label string, value string) string {
	if value[0:4] != "gsm:" {
		log.Debugf("[Config] Not reading %s from GSM", label)
		return value
	}
	name := value[4:]
	log.Debugf("[Config] Reading %s from GSM for %s", name, label)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(context.Background(), req)
	if err != nil {
		log.Errorf("[Config] Failed to access secret %s: %v", name, err)
		return ""
	}

	log.Debugf("[Config] Successfully read %s from GSM for %s", name, label)
	return string(result.Payload.Data)
}

func LoadSecretsFromGSM(config models.Config) models.Config {
	log.Debugf("[Config] Loading secrets from GSM")
	configWithSecrets := config

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Errorf("[Config] Failed to create secretmanager client: %v", err)
		return configWithSecrets
	}
	defer client.Close()

	configWithSecrets.MongoDB.URI = readSecretFromGSM(client, "MongoDB.URI", config.MongoDB.URI)

	for i, ethNetwork := range config.EthereumNetworks {
		configWithSecrets.EthereumNetworks[i].PrivateKey = readSecretFromGSM(client, fmt.Sprintf("EthereumNetworks.[%d].PrivateKey", i), ethNetwork.PrivateKey)
	}

	for i, cosmosNetwork := range config.CosmosNetworks {
		configWithSecrets.CosmosNetworks[i].PrivateKey = readSecretFromGSM(client, fmt.Sprintf("CosmosNetworks.[%d].PrivateKey", i), cosmosNetwork.PrivateKey)
	}

	log.Debugf("[Config] Successfully loaded secrets from GSM")
	return configWithSecrets
}
