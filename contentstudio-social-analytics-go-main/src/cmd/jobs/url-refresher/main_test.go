package main

import (
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

func TestBuildRateManager_Defaults(t *testing.T) {
	orig := newRateManager
	defer func() { newRateManager = orig }()

	var captured social.RateLimits
	newRateManager = func(rl social.RateLimits) *social.RateManager {
		captured = rl
		return &social.RateManager{}
	}

	cfg := &config.Config{}
	log := logger.New("info")

	if rm := buildRateManager(cfg, *log); rm == nil {
		t.Fatal("expected non-nil rate manager")
	}

	if captured.PerTokenRPS != 3.0 || captured.PerTokenBurst != 3 || captured.GlobalRPS != 10.0 || captured.GlobalBurst != 10 {
		t.Fatalf("unexpected defaults: %+v", captured)
	}
}

func TestBuildRateManager_UsesConfiguredValues(t *testing.T) {
	orig := newRateManager
	defer func() { newRateManager = orig }()

	var captured social.RateLimits
	newRateManager = func(rl social.RateLimits) *social.RateManager {
		captured = rl
		return &social.RateManager{}
	}

	cfg := &config.Config{
		Facebook: config.FacebookConfig{
			PerTokenRPS:   7.5,
			PerTokenBurst: 8,
			GlobalRPS:     21.0,
			GlobalBurst:   22,
		},
	}
	log := logger.New("info")

	_ = buildRateManager(cfg, *log)

	if captured.PerTokenRPS != 7.5 || captured.PerTokenBurst != 8 || captured.GlobalRPS != 21.0 || captured.GlobalBurst != 22 {
		t.Fatalf("unexpected configured values: %+v", captured)
	}
}
