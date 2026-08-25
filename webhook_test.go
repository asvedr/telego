package telego

import (
	"testing"
	"time"

	"gitlab.com/asvedr/testify/assert"
	"gitlab.com/asvedr/testify/mock"

	"github.com/asvedr/telego/internal/json"
	ta "github.com/asvedr/telego/telegoapi"
)

func TestBot_UpdatesViaWebhook(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		b, err := NewBot(validToken, WithDiscardLogger())
		assert.NoError(t, err)

		_, err = b.UpdatesViaWebhook(t.Context(), func(handler WebhookHandler) error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("error_webhook_exist", func(t *testing.T) {
		b := &Bot{}

		_, err := b.UpdatesViaWebhook(t.Context(), func(handler WebhookHandler) error {
			return nil
		})
		assert.NoError(t, err)

		_, err = b.UpdatesViaWebhook(t.Context(), func(handler WebhookHandler) error {
			return nil
		})
		assert.Error(t, err)
	})

	t.Run("error_long_polling_exist", func(t *testing.T) {
		m := newMockedBot()

		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(data, nil).Maybe()

		resp := telegoResponse(t, []Update{
			{UpdateID: 1},
			{UpdateID: 2},
		})
		m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return(resp, nil).Maybe()

		_, err := m.Bot.UpdatesViaLongPolling(t.Context(), nil)
		assert.NoError(t, err)

		_, err = m.Bot.UpdatesViaWebhook(t.Context(), func(handler WebhookHandler) error {
			return nil
		})
		assert.Error(t, err)
	})

	t.Run("end_to_end", func(t *testing.T) {
		b, err := NewBot(validToken, WithDiscardLogger())
		assert.NoError(t, err)

		pushUpdate := make(chan struct{})

		expectedUpdate := Update{
			UpdateID: 1,
			Message:  &Message{Text: "ok"},
		}
		expectedUpdateBytes, err := json.Marshal(expectedUpdate)
		assert.NoError(t, err)

		updates, err := b.UpdatesViaWebhook(t.Context(), func(handler WebhookHandler) error {
			go func() {
				<-pushUpdate
				err = handler(t.Context(), expectedUpdateBytes)
				assert.NoError(t, err)
			}()
			return nil
		})
		assert.NoError(t, err)

		pushUpdate <- struct{}{}

		select {
		case update, ok := <-updates:
			assert.True(t, ok)
			update.ctx = nil
			assert.Equal(t, expectedUpdate, update)
		case <-time.After(timeout):
			t.Fatalf("Timeout")
		}
	})
}

func TestWithWebhookBuffer(t *testing.T) {
	ctx := &webhook{}
	buffer := uint(1)

	err := WithWebhookBuffer(buffer)(nil, ctx)
	assert.NoError(t, err)
	assert.Equal(t, buffer, ctx.updateChanBuffer)
}

func TestWithWebhookSet(t *testing.T) {
	m := newMockedBot()
	ctx := &webhook{}

	m.MockRequestConstructor.On("JSONRequest", mock.Anything).Return(&ta.RequestData{
		BodyRaw: []byte{},
	}, nil)

	m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).Return(&ta.Response{Ok: true}, nil)

	err := WithWebhookSet(t.Context(), &SetWebhookParams{})(m.Bot, ctx)
	assert.NoError(t, err)
}
