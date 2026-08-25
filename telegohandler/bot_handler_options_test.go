package telegohandler

import (
	"testing"

	"gitlab.com/asvedr/testify/assert"

	"github.com/asvedr/telego"
)

func TestWithErrorHandler(t *testing.T) {
	bh := &BotHandler{}
	handler := func(ctx *Context, update telego.Update, err error) {}

	err := WithErrorHandler(handler)(bh)
	assert.NoError(t, err)
}
