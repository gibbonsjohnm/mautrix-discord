package main

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/id"

	"go.mau.fi/mautrix-discord/config"
)

func TestNormalizeBotToken(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"   ":           "",
		"abc.def":       "Bot abc.def",
		"Bot abc.def":   "Bot abc.def",
		"  abc.def\n":   "Bot abc.def",
		"Bot abc.def  ": "Bot abc.def",
	}
	for in, want := range cases {
		if got := normalizeBotToken(in); got != want {
			t.Errorf("normalizeBotToken(%q) = %q, want %q", in, got, want)
		}
	}
}

func newTestBridge(bots map[string]string) *DiscordBridge {
	br := &DiscordBridge{Config: &config.Config{}}
	br.Config.Bridge.ReactionRelayBots = bots
	nop := zerolog.Nop()
	br.ZLog = &nop
	return br
}

func TestInitReactionRelayBots(t *testing.T) {
	orig := discordUserLookup
	t.Cleanup(func() { discordUserLookup = orig })
	discordUserLookup = func(sess *discordgo.Session) (string, error) {
		switch sess.Token {
		case "Bot good":
			return "111", nil
		case "Bot bad":
			return "", errors.New("401 unauthorized")
		}
		return "", errors.New("unexpected token " + sess.Token)
	}

	br := newTestBridge(map[string]string{
		"@relay:example.com":  "good",
		"@broken:example.com": "bad",
		"@empty:example.com":  "",
	})
	br.initReactionRelayBots()

	bot := br.GetReactionRelayBot(id.UserID("@relay:example.com"))
	if bot == nil {
		t.Fatal("expected relay bot for @relay:example.com")
	}
	if bot.UserID != "111" {
		t.Errorf("bot.UserID = %q, want 111", bot.UserID)
	}
	if bot.Session == nil || bot.Session.Token != "Bot good" {
		t.Errorf("bot session not built with normalized token")
	}
	if !br.IsReactionRelayBot("111") {
		t.Error("IsReactionRelayBot(111) = false, want true")
	}
	if br.GetReactionRelayBot(id.UserID("@broken:example.com")) != nil {
		t.Error("bot with failing auth should be skipped")
	}
	if br.GetReactionRelayBot(id.UserID("@empty:example.com")) != nil {
		t.Error("bot with empty token should be skipped")
	}
	if br.IsReactionRelayBot("999") {
		t.Error("IsReactionRelayBot(999) = true, want false")
	}
}

func TestReactionRelayBotsUnconfigured(t *testing.T) {
	br := &DiscordBridge{}
	if br.GetReactionRelayBot(id.UserID("@x:example.com")) != nil || br.IsReactionRelayBot("1") {
		t.Error("nil registry must behave as empty")
	}
}
