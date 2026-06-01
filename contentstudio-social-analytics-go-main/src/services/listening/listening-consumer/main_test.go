package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

func TestMongoClientOptionsWithoutCredentialsDoesNotSetAuth(t *testing.T) {
	cfg := &config.Config{
		Mongo: config.MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "contentstudiobackend",
		},
	}

	opts := mongoClientOptions(cfg)

	require.NotNil(t, opts)
	assert.Nil(t, opts.Auth)
}

func TestMongoClientOptionsWithCredentialsSetsAuth(t *testing.T) {
	cfg := &config.Config{
		Mongo: config.MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "contentstudiobackend",
			Username: "listener",
			Password: "secret",
		},
	}

	opts := mongoClientOptions(cfg)

	require.NotNil(t, opts)
	require.NotNil(t, opts.Auth)
	assert.Equal(t, "listener", opts.Auth.Username)
	assert.Equal(t, "secret", opts.Auth.Password)
	assert.Equal(t, "contentstudiobackend", opts.Auth.AuthSource)
}
