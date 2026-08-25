package telego

import (
	"testing"
	"time"

	"gitlab.com/asvedr/testify/assert"
	"gitlab.com/asvedr/testify/mock"
)

const timeout = time.Second

func TestBot_UpdatesViaLongPolling(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := newMockedBot()

		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(data, nil)

		expectedUpdates := []Update{
			{UpdateID: 1},
			{UpdateID: 2},
		}
		resp := telegoResponse(t, expectedUpdates)
		m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return(resp, nil)

		assert.NotPanics(t, func() {
			updates, err := m.Bot.UpdatesViaLongPolling(t.Context(), nil)
			assert.NoError(t, err)

			time.Sleep(time.Millisecond * 10)
			select {
			case <-time.After(timeout):
				t.Fatal("Timeout")
			case update := <-updates:
				assert.NotZero(t, update.UpdateID)
			}
		})
	})

	t.Run("error_get_update", func(t *testing.T) {
		m := newMockedBot()

		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(nil, errTest)

		assert.NotPanics(t, func() {
			_, err := m.Bot.UpdatesViaLongPolling(t.Context(), nil)
			assert.NoError(t, err)
			time.Sleep(time.Millisecond * 10)
		})
	})

	t.Run("error_already_running", func(t *testing.T) {
		m := newMockedBot()

		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(nil, errTest).Maybe()

		assert.NotPanics(t, func() {
			_, err := m.Bot.UpdatesViaLongPolling(t.Context(), nil)
			assert.NoError(t, err)

			_, err = m.Bot.UpdatesViaLongPolling(t.Context(), nil)
			assert.Error(t, err)
		})
	})

	t.Run("error_options", func(t *testing.T) {
		m := newMockedBot()

		assert.NotPanics(t, func() {
			_, err := m.Bot.UpdatesViaLongPolling(t.Context(), nil, WithLongPollingUpdateInterval(-time.Second))
			assert.Error(t, err)
		})
	})
}

func TestWithLongPollingUpdateInterval(t *testing.T) {
	ctx := &longPolling{}
	interval := time.Second

	t.Run("success", func(t *testing.T) {
		err := WithLongPollingUpdateInterval(interval)(ctx)
		assert.NoError(t, err)
		assert.Equal(t, interval, ctx.updateInterval)
	})

	t.Run("error", func(t *testing.T) {
		err := WithLongPollingUpdateInterval(-interval)(ctx)
		assert.Error(t, err)
	})
}

func TestWithLongPollingRetryTimeout(t *testing.T) {
	ctx := &longPolling{}

	t.Run("success", func(t *testing.T) {
		err := WithLongPollingRetryTimeout(timeout)(ctx)
		assert.NoError(t, err)
		assert.Equal(t, timeout, ctx.retryTimeout)
	})

	t.Run("error", func(t *testing.T) {
		err := WithLongPollingRetryTimeout(-timeout)(ctx)
		assert.Error(t, err)
	})
}

func TestWithLongPollingBuffer(t *testing.T) {
	ctx := &longPolling{}
	buffer := uint(1)

	err := WithLongPollingBuffer(buffer)(ctx)
	assert.NoError(t, err)
	assert.Equal(t, buffer, ctx.updateChanBuffer)
}
