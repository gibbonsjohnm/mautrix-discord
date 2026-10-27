// mautrix-discord - A Matrix-Discord puppeting bridge.
// Copyright (C) 2022 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"maunium.net/go/mautrix/id"
)

// ReactionRelayBot is a Discord bot account that stands in for a single
// relay-only Matrix user. Discord webhooks cannot add reactions, so reactions
// from Matrix users without a Discord login are relayed through this bot.
type ReactionRelayBot struct {
	MXID    id.UserID
	UserID  string
	Session *discordgo.Session
}

type reactionRelayRegistry struct {
	byMXID   map[id.UserID]*ReactionRelayBot
	byUserID map[string]*ReactionRelayBot
	lock     sync.RWMutex
}

func newReactionRelayRegistry() *reactionRelayRegistry {
	return &reactionRelayRegistry{
		byMXID:   make(map[id.UserID]*ReactionRelayBot),
		byUserID: make(map[string]*ReactionRelayBot),
	}
}

func (r *reactionRelayRegistry) add(bot *ReactionRelayBot) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.byMXID[bot.MXID] = bot
	r.byUserID[bot.UserID] = bot
}

func (r *reactionRelayRegistry) getByMXID(mxid id.UserID) *ReactionRelayBot {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.byMXID[mxid]
}

func (r *reactionRelayRegistry) getByUserID(userID string) *ReactionRelayBot {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.byUserID[userID]
}

// normalizeBotToken makes sure a configured token carries the "Bot " prefix
// that Discord requires for bot accounts, so operators can paste the raw token
// from the developer portal.
func normalizeBotToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if strings.HasPrefix(token, "Bot ") {
		return token
	}
	return "Bot " + token
}

// discordUserLookup resolves a bot token to its Discord user ID. It is a
// variable so tests can avoid hitting the network.
var discordUserLookup = func(sess *discordgo.Session) (string, error) {
	me, err := sess.User("@me")
	if err != nil {
		return "", err
	}
	return me.ID, nil
}

// initReactionRelayBots builds a REST-only Discord session for every entry in
// bridge.reaction_relay_bots. A bot that fails to authenticate is logged and
// skipped so a bad token never stops the bridge from starting.
func (br *DiscordBridge) initReactionRelayBots() {
	br.reactionRelayBots = newReactionRelayRegistry()
	for mxidStr, rawToken := range br.Config.Bridge.ReactionRelayBots {
		mxid := id.UserID(mxidStr)
		log := br.ZLog.With().Str("component", "reaction_relay").Str("mxid", mxid.String()).Logger()
		token := normalizeBotToken(rawToken)
		if token == "" {
			log.Warn().Msg("Skipping reaction relay bot with empty token")
			continue
		}
		sess, err := discordgo.New(token)
		if err != nil {
			log.Err(err).Msg("Failed to create Discord session for reaction relay bot")
			continue
		}
		userID, err := discordUserLookup(sess)
		if err != nil {
			log.Err(err).Msg("Failed to authenticate reaction relay bot, reactions from this user will not be relayed")
			continue
		}
		br.reactionRelayBots.add(&ReactionRelayBot{MXID: mxid, UserID: userID, Session: sess})
		log.Info().Str("bot_user_id", userID).Msg("Reaction relay bot ready")
	}
}

// GetReactionRelayBot returns the stand-in bot for a Matrix user, or nil.
func (br *DiscordBridge) GetReactionRelayBot(mxid id.UserID) *ReactionRelayBot {
	if br.reactionRelayBots == nil {
		return nil
	}
	return br.reactionRelayBots.getByMXID(mxid)
}

// IsReactionRelayBot reports whether a Discord user ID belongs to one of the
// configured stand-in bots.
func (br *DiscordBridge) IsReactionRelayBot(userID string) bool {
	if br.reactionRelayBots == nil {
		return false
	}
	return br.reactionRelayBots.getByUserID(userID) != nil
}
