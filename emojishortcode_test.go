package main

import (
	"strings"
	"testing"
)

func TestParseEmojiShortcode(t *testing.T) {
	cases := []struct {
		in   string
		name string
		ok   bool
	}{
		{":kekw:", "kekw", true},
		{":Pepe_Laugh2:", "Pepe_Laugh2", true},
		{"kekw", "", false},
		{":kekw", "", false},
		{"::", "", false},
		{":a:", "", false},
		{":has space:", "", false},
		{":has-dash:", "", false},
		{"🎉", "", false},
		{"mxc://example.org/abc", "", false},
		{":" + strings.Repeat("a", 32) + ":", strings.Repeat("a", 32), true},
		{":" + strings.Repeat("a", 33) + ":", "", false},
	}
	for _, c := range cases {
		name, ok := parseEmojiShortcode(c.in)
		if ok != c.ok || name != c.name {
			t.Errorf("parseEmojiShortcode(%q) = (%q, %v), want (%q, %v)", c.in, name, ok, c.name, c.ok)
		}
	}
}
