package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.mau.fi/util/configupgrade"
	"gopkg.in/yaml.v3"

	"go.mau.fi/mautrix-discord/config"
)

// Regression test: the example config must declare reaction_relay_bots as an
// (empty) mapping, otherwise the config upgrader can't splice a user-provided
// map into it and the upgraded document fails to unmarshal with
// "cannot unmarshal !!map into map[string]string".
func TestConfigUpgradeKeepsReactionRelayBots(t *testing.T) {
	userCfg := strings.Replace(ExampleConfig,
		"reaction_relay_bots: {}",
		"reaction_relay_bots:\n        '@relay:example.org': 'Bot abc.def.ghi'\n", 1)
	if userCfg == ExampleConfig {
		t.Fatal("example config no longer contains `reaction_relay_bots: {}`")
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(userCfg), 0o600); err != nil {
		t.Fatal(err)
	}

	upgrader := &configupgrade.StructUpgrader{
		SimpleUpgrader: configupgrade.SimpleUpgrader(config.DoUpgrade),
		Blocks:         config.SpacedBlocks,
		Base:           ExampleConfig,
	}
	upgraded, _, err := configupgrade.Do(path, false, upgrader)
	if err != nil {
		t.Fatalf("config upgrade failed: %v", err)
	}

	var cfg config.Config
	if err = yaml.Unmarshal(upgraded, &cfg); err != nil {
		t.Fatalf("upgraded config does not unmarshal: %v", err)
	}
	if got := cfg.Bridge.ReactionRelayBots["@relay:example.org"]; got != "Bot abc.def.ghi" {
		t.Fatalf("reaction_relay_bots not preserved through upgrade, got %q", got)
	}
}
