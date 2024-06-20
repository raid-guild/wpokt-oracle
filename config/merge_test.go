package config

import (
	"testing"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
)

func TestMergeConfigs(t *testing.T) {
	yamlConfig := models.Config{
		MongoDB: models.MongoConfig{
			URI:       "mongodb://localhost:27017",
			Database:  "testdb",
			TimeoutMS: 1000,
		},
	}

	envConfig := models.Config{
		MongoDB: models.MongoConfig{
			URI: "mongodb://localhost:27018",
		},
	}

	mergedConfig := mergeConfigs(yamlConfig, envConfig)
	assert.Equal(t, "mongodb://localhost:27018", mergedConfig.MongoDB.URI)
	assert.Equal(t, "testdb", mergedConfig.MongoDB.Database)
	assert.Equal(t, uint64(1000), mergedConfig.MongoDB.TimeoutMS)
}
