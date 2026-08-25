//go:build integration && interactive

package main

import (
	"encoding/json"
	"testing"

	"gitlab.com/asvedr/testify/assert"

	"github.com/asvedr/telego"
	th "github.com/asvedr/telego/telegohandler"
)

func TestGetMessage(t *testing.T) {
	ctx := t.Context()

	updates, err := bot.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		AllowedUpdates: []string{
			telego.MessageUpdates,
		},
	})
	assert.NoError(t, err)

	bh, err := th.NewBotHandler(bot, updates)
	assert.NoError(t, err)

	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		var data []byte
		data, err = json.Marshal(message)
		assert.NoError(t, err)

		t.Log(string(data))

		return nil
	})

	err = bh.Start()
	assert.NoError(t, err)
}
