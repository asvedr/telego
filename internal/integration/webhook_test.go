//go:build integration && webhook

package main

import (
	"testing"

	"gitlab.com/asvedr/testify/assert"

	tu "github.com/mymmrac/telego/telegoutil"
)

func TestWebhookInfo(t *testing.T) {
	ctx := t.Context()

	info, err := bot.GetWebhookInfo(ctx)
	assert.NoError(t, err)

	t.Logf("WebhookInfo: %+v", info)
}

func TestWebhook(t *testing.T) {
	ctx := t.Context()

	err := bot.SetWebhook(ctx, tu.Webhook("https://example.org"))
	assert.NoError(t, err)
}

func TestDeleteWebhook(t *testing.T) {
	ctx := t.Context()

	err := bot.DeleteWebhook(ctx, nil)
	assert.NoError(t, err)
}
