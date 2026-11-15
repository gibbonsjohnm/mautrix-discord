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

import "regexp"

// Discord custom emoji names are 2-32 characters of letters, digits and underscores.
var emojiShortcodeRegex = regexp.MustCompile(`^:([A-Za-z0-9_]{2,32}):$`)

// parseEmojiShortcode returns the emoji name if key has the ":name:" form the
// bridge uses for custom emoji reactions when custom_emoji_reactions is false.
func parseEmojiShortcode(key string) (string, bool) {
	m := emojiShortcodeRegex.FindStringSubmatch(key)
	if m == nil {
		return "", false
	}
	return m[1], true
}
