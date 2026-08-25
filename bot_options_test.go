package telego

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/valyala/fasthttp"
	"gitlab.com/asvedr/testify/assert"

	"github.com/asvedr/telego/internal/json"
	ta "github.com/asvedr/telego/telegoapi"
	mockapi "github.com/asvedr/telego/telegoapi/mock"
)

type testCallerType struct{}

func (c testCallerType) Call(_ context.Context, _ string, _ *ta.RequestData) (*ta.Response, error) {
	panic("unreachable: testCallerType.Call")
}

func TestWithAPICaller(t *testing.T) {
	bot := &Bot{}
	caller := testCallerType{}

	err := WithAPICaller(caller)(bot)
	assert.NoError(t, err)
	assert.EqualValues(t, caller, bot.api)
}

func TestWithFastHTTPClient(t *testing.T) {
	bot := &Bot{}
	client := &fasthttp.Client{}

	err := WithFastHTTPClient(client)(bot)
	assert.NoError(t, err)
}

func TestWithHTTPClient(t *testing.T) {
	bot := &Bot{}
	client := &http.Client{}

	err := WithHTTPClient(client)(bot)
	assert.NoError(t, err)
}

type testConstructorType struct{}

func (testConstructorType) JSONRequest(_ any) (*ta.RequestData, error) {
	panic("unreachable: testConstructorType.JSONRequest")
}

func (testConstructorType) MultipartRequest(
	_ map[string]string, _ map[string]ta.NamedReader,
) (*ta.RequestData, error) {
	panic("unreachable: testConstructorType.MultipartRequest")
}

func TestWithRequestConstructor(t *testing.T) {
	bot := &Bot{}
	constructor := &testConstructorType{}

	err := WithRequestConstructor(constructor)(bot)
	assert.NoError(t, err)
	assert.EqualValues(t, constructor, bot.constructor)
}

func TestWithDefaultLogger(t *testing.T) {
	bot := &Bot{}

	err := WithDefaultLogger(true, false)(bot)
	assert.NoError(t, err)

	log, ok := bot.log.(*logger)
	assert.True(t, ok)

	assert.True(t, log.DebugMode)
	assert.False(t, log.PrintErrors)
	assert.NotNil(t, log.Replacer)
}

func TestWithExtendedDefaultLogger(t *testing.T) {
	bot := &Bot{}

	t.Run("nil_replacer", func(t *testing.T) {
		err := WithExtendedDefaultLogger(true, true, nil)(bot)
		assert.NoError(t, err)

		log, ok := bot.log.(*logger)
		assert.True(t, ok)

		assert.True(t, log.DebugMode)
		assert.True(t, log.PrintErrors)
		assert.Nil(t, log.Replacer)
	})

	t.Run("not_nil_replacer", func(t *testing.T) {
		err := WithExtendedDefaultLogger(true, true, strings.NewReplacer("old", "new"))(bot)
		assert.NoError(t, err)

		log, ok := bot.log.(*logger)
		assert.True(t, ok)

		assert.True(t, log.DebugMode)
		assert.True(t, log.PrintErrors)
		assert.NotNil(t, log.Replacer)
	})
}

func TestWithDiscardLogger(t *testing.T) {
	bot := &Bot{}

	err := WithDiscardLogger()(bot)
	assert.NoError(t, err)

	log, ok := bot.log.(*logger)
	assert.True(t, ok)

	assert.False(t, log.DebugMode)
	assert.False(t, log.PrintErrors)
	assert.NotNil(t, log.Replacer)
}

type testLoggerType struct{}

func (testLoggerType) Debugf(_ string, _ ...any) {
	// NoOp
}

func (testLoggerType) Errorf(_ string, _ ...any) {
	// NoOp
}

func TestWithLogger(t *testing.T) {
	bot := &Bot{}
	log := &testLoggerType{}

	err := WithLogger(log)(bot)
	assert.NoError(t, err)
	assert.EqualValues(t, log, bot.log)
}

func TestWithAPIServer(t *testing.T) {
	bot := &Bot{}

	t.Run("success", func(t *testing.T) {
		err := WithAPIServer("test")(bot)
		assert.NoError(t, err)
		assert.Equal(t, "test", bot.apiURL)
	})

	t.Run("error", func(t *testing.T) {
		err := WithAPIServer("")(bot)
		assert.Error(t, err)
	})
}

func TestWithDefaultDebugLogger(t *testing.T) {
	bot := &Bot{}

	err := WithDefaultDebugLogger()(bot)
	assert.NoError(t, err)

	log, ok := bot.log.(*logger)
	assert.True(t, ok)

	assert.True(t, log.DebugMode)
	assert.True(t, log.PrintErrors)
	assert.NotNil(t, log.Replacer)
}

func TestWithDebugMode(t *testing.T) {
	bot := &Bot{}

	err := WithDebugMode()(bot)
	assert.NoError(t, err)

	assert.True(t, bot.debugMode)
}

func TestWithTestServerPath(t *testing.T) {
	bot := &Bot{}

	err := WithTestServerPath()(bot)
	assert.NoError(t, err)

	assert.True(t, bot.useTestServerPath)
}

func TestWithHealthCheck(t *testing.T) {
	caller := mockapi.NewMockCaller()
	constructor := mockapi.NewMockRequestConstructor()

	expectedResp := &ta.Response{
		Ok:     true,
		Result: json.RawMessage(`{}`),
	}

	expectedData := &ta.RequestData{
		ContentType: ta.ContentTypeJSON,
		BodyRaw:     []byte{},
	}

	constructor.On("JSONRequest", nil).
		Return(expectedData, nil).
		Times(1)

	caller.On("Call", t.Context(), defaultBotAPIServer+botPathPrefix+validToken+"/getMe", expectedData).
		Return(expectedResp, nil).
		Times(1)

	bot, err := NewBot(validToken,
		WithAPICaller(caller),
		WithRequestConstructor(constructor),
		WithHealthCheck(t.Context()),
	)
	assert.NoError(t, err)
	assert.NotNil(t, bot)
}

func TestWithWarnings(t *testing.T) {
	bot := &Bot{}

	err := WithWarnings()(bot)
	assert.NoError(t, err)

	assert.True(t, bot.reportWarningAsErrors)
}
