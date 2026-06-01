package main

import (
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

func TestSchedulerInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seconds int
		want    time.Duration
	}{
		{"zero uses default", 0, defaultSchedulerInterval},
		{"negative uses default", -1, defaultSchedulerInterval},
		{"30 seconds", 30, 30 * time.Second},
		{"60 seconds is one minute", 60, time.Minute},
		{"3600 seconds is one hour", 3600, time.Hour},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := schedulerInterval(tc.seconds)
			if got != tc.want {
				t.Errorf("schedulerInterval(%d) = %v, want %v", tc.seconds, got, tc.want)
			}
		})
	}
}

func TestMongoClientOptions_NoAuth(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Mongo: config.MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "testdb",
		},
	}

	opts := mongoClientOptions(cfg)
	if opts == nil {
		t.Fatal("expected non-nil options")
	}
	if opts.Auth != nil {
		t.Errorf("expected no auth to be set, got %+v", opts.Auth)
	}
}

func TestMongoClientOptions_WithAuth(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Mongo: config.MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "testdb",
			Username: "user",
			Password: "pass",
		},
	}

	opts := mongoClientOptions(cfg)
	if opts == nil {
		t.Fatal("expected non-nil options")
	}
	if opts.Auth == nil {
		t.Fatal("expected auth to be set")
	}
	if opts.Auth.Username != "user" {
		t.Errorf("username: want %q, got %q", "user", opts.Auth.Username)
	}
	if opts.Auth.Password != "pass" {
		t.Errorf("password: want %q, got %q", "pass", opts.Auth.Password)
	}
	if opts.Auth.AuthSource != "testdb" {
		t.Errorf("auth_source: want %q, got %q", "testdb", opts.Auth.AuthSource)
	}
}
