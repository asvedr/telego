//go:build integration && interactive

package main

import (
	"fmt"
	"testing"

	"gitlab.com/asvedr/testify/assert"

	"github.com/asvedr/telego"
	th "github.com/asvedr/telego/telegohandler"
	tu "github.com/asvedr/telego/telegoutil"
)

func TestInlineQuery(t *testing.T) {
	ctx := t.Context()

	updates, err := bot.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		AllowedUpdates: []string{
			telego.InlineQueryUpdates,
			telego.ChosenInlineResultUpdates,
		},
	})
	assert.NoError(t, err)

	bh, err := th.NewBotHandler(bot, updates)
	assert.NoError(t, err)

	bh.HandleInlineQuery(func(ctx *th.Context, query telego.InlineQuery) error {
		t.Log(query.Query)

		err = ctx.Bot().AnswerInlineQuery(ctx, &telego.AnswerInlineQueryParams{
			InlineQueryID: query.ID,
			Results: []telego.InlineQueryResult{
				tu.ResultArticle("1", "Echo", tu.TextMessage("["+query.Query+"]")).WithDescription(query.Query),
			},
			CacheTime:  1,
			IsPersonal: true,
		})
		if err != nil {
			return fmt.Errorf("answer inline query: %w", err)
		}

		return nil
	})

	bh.HandleChosenInlineResult(func(ctx *th.Context, result telego.ChosenInlineResult) error {
		t.Log(result.Query, result.ResultID)
		return nil
	})

	err = bh.Start()
	assert.NoError(t, err)
}
