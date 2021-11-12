package main

import (
	"testing"

	"github.com/matryer/xbar/pkg/plugins"
)

func TestEmoji(t *testing.T) {
	em := ListEmoji()
	for _, e := range em {
		t.Logf("listed: '%s'", e)
		_, err := GetEmoji(e)
		if err != nil {
			t.Error(err.Error())
			return
		}
		emojise := plugins.Emojize(":" + e + ":")
		t.Logf("emojiised: '%s'", emojise)
	}
}
