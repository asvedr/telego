package telego

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"gitlab.com/asvedr/testify/assert"
	"gitlab.com/asvedr/testify/mock"

	"github.com/mymmrac/telego/internal/json"
	ta "github.com/mymmrac/telego/telegoapi"
	mockapi "github.com/mymmrac/telego/telegoapi/mock"
)

const (
	validToken   = "1234567890:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc"
	invalidToken = "invalid-token"

	methodName = "testMethod"
)

var errTest = errors.New("error")

func Test_validateToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		isValid bool
	}{
		{
			name:    "empty",
			token:   "",
			isValid: false,
		},
		{
			name:    "not_valid",
			token:   invalidToken,
			isValid: false,
		},
		{
			name:    "valid_1",
			token:   validToken,
			isValid: true,
		},
		{
			name:    "valid_2",
			token:   "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc",
			isValid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateToken(tt.token)
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestNewBot(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bot, err := NewBot(validToken)

		assert.NoError(t, err)
		assert.NotNil(t, bot)
	})

	t.Run("success_with_options", func(t *testing.T) {
		bot, err := NewBot(validToken, func(_ *Bot) error { return nil })

		assert.NoError(t, err)
		assert.NotNil(t, bot)
	})

	t.Run("error", func(t *testing.T) {
		bot, err := NewBot(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, bot)
	})

	t.Run("error_with_options", func(t *testing.T) {
		bot, err := NewBot(validToken, func(_ *Bot) error { return errTest })

		assert.ErrorIs(t, err, errTest)
		assert.Nil(t, bot)
	})

	t.Run("with_health_check", func(t *testing.T) {
		caller := mockapi.NewMockCaller()
		constructor := mockapi.NewMockRequestConstructor()

		expectedData := &ta.RequestData{
			ContentType: ta.ContentTypeJSON,
			BodyRaw:     []byte{},
		}

		t.Run("success", func(t *testing.T) {
			expectedResp := &ta.Response{
				Ok:     true,
				Result: json.RawMessage(`{}`),
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
		})

		t.Run("error", func(t *testing.T) {
			expectedResp := &ta.Response{
				Ok:    false,
				Error: &ta.Error{},
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

			assert.Error(t, err)
			assert.Nil(t, bot)
		})
	})
}

func TestBot_Token(t *testing.T) {
	bot, err := NewBot(validToken)
	assert.NoError(t, err)

	assert.Equal(t, validToken, bot.Token())
}

func TestBot_SecretToken(t *testing.T) {
	bot, err := NewBot(validToken)
	assert.NoError(t, err)

	hash := sha256.Sum256([]byte(validToken))
	assert.Equal(t, hex.EncodeToString(hash[:]), bot.SecretToken())
}

func TestBot_Logger(t *testing.T) {
	bot, err := NewBot(validToken)
	assert.NoError(t, err)

	assert.Equal(t, bot.log, bot.Logger())
}

func TestBot_FileDownloadURL(t *testing.T) {
	t.Run("regular", func(t *testing.T) {
		bot, err := NewBot(validToken)
		assert.NoError(t, err)

		filepath := "file.txt"
		url := bot.FileDownloadURL(filepath)
		assert.Equal(t, bot.apiURL+"/file"+botPathPrefix+bot.token+"/"+filepath, url)
	})

	t.Run("test", func(t *testing.T) {
		bot, err := NewBot(validToken, WithTestServerPath())
		assert.NoError(t, err)

		filepath := "file.txt"
		url := bot.FileDownloadURL(filepath)
		assert.Equal(t, bot.apiURL+"/file"+botPathPrefix+bot.token+"/test/"+filepath, url)
	})
}

type testErrorMarshal struct {
	Number int `json:"number"`
}

func (t testErrorMarshal) MarshalJSON() ([]byte, error) {
	return nil, errTest
}

type testEmptyMarshal struct {
	Number int `json:"number"`
}

func (t testEmptyMarshal) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}

func Test_parseParameters(t *testing.T) {
	n := 1

	tests := []struct {
		name             string
		parameters       any
		parsedParameters map[string]string
		isError          bool
	}{
		{
			name: "success",
			parameters: &struct {
				Empty       string    `json:"empty,omitempty"`
				EmptyNoOmit string    `json:"empty_no_omit"`
				Number      int       `json:"number"`
				Array       []int     `json:"array"`
				Text        string    `json:"text"`
				Struct      *struct { //revive:disable:nested-structs
					N int `json:"n"`
				} `json:"struct"`
			}{
				Number: 10,
				Array:  []int{1, 2, 3},
				Text:   "ok",
				Struct: &struct {
					N int `json:"n"`
				}{2},
			},
			parsedParameters: map[string]string{
				"number": "10",
				"array":  "[1,2,3]",
				"struct": "{\"n\":2}",
				"text":   "ok",
			},
			isError: false,
		},
		{
			name: "error_not_pointer",
			parameters: struct {
				a int
			}{},
			parsedParameters: nil,
			isError:          true,
		},
		{
			name:             "error_not_struct",
			parameters:       &n,
			parsedParameters: nil,
			isError:          true,
		},
		{
			name: "error_no_tag",
			parameters: &struct {
				Number int
			}{
				Number: 1,
			},
			parsedParameters: nil,
			isError:          true,
		},
		{
			name: "error_marshal",
			parameters: &struct {
				NonMarshaled testErrorMarshal `json:"non_marshaled"`
			}{
				NonMarshaled: testErrorMarshal{1},
			},
			parsedParameters: nil,
			isError:          true,
		},
		{
			name: "success_get_update",
			parameters: &GetUpdatesParams{
				Offset:         1,
				Limit:          2,
				Timeout:        3,
				AllowedUpdates: []string{"ok1", "ok2"},
			},
			parsedParameters: map[string]string{
				"offset":          "1",
				"limit":           "2",
				"timeout":         "3",
				"allowed_updates": "[\"ok1\",\"ok2\"]",
			},
			isError: false,
		},
		{
			name: "success_send_photo",
			parameters: &SendPhotoParams{
				ChatID:              ChatID{ID: 1},
				Photo:               InputFile{URL: "ok1"},
				Caption:             "ok2",
				DisableNotification: true,
				ReplyMarkup: &InlineKeyboardMarkup{
					InlineKeyboard: [][]InlineKeyboardButton{
						{
							{
								Text: "ok3",
							},
						},
					},
				},
			},
			parsedParameters: map[string]string{
				"caption":              "ok2",
				"chat_id":              "1",
				"disable_notification": "true",
				"photo":                "ok1",
				"reply_markup":         "{\"inline_keyboard\":[[{\"text\":\"ok3\"}]]}",
			},
			isError: false,
		},
		{
			name: "success_empty_marshal",
			parameters: &struct {
				EmptyMarshaled testEmptyMarshal `json:"empty_marshaled"`
			}{
				EmptyMarshaled: testEmptyMarshal{1},
			},
			parsedParameters: map[string]string{},
			isError:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedParameters, err := parseParameters(tt.parameters)
			if tt.isError {
				assert.Error(t, err)
				assert.Nil(t, parsedParameters)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.parsedParameters, parsedParameters)
		})
	}
}

type testStruct struct{}

func (ts *testStruct) fileParameters() map[string]ta.NamedReader {
	return map[string]ta.NamedReader{
		"test": &testNamedReader{},
	}
}

func Test_filesParameters(t *testing.T) {
	tests := []struct {
		name       string
		parameters any
		files      map[string]ta.NamedReader
		hasFiles   bool
	}{
		{
			name:       "with_files",
			parameters: &testStruct{},
			files: map[string]ta.NamedReader{
				"test": &testNamedReader{},
			},
			hasFiles: true,
		},
		{
			name:       "no_files",
			parameters: 1,
			files:      nil,
			hasFiles:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, hasFiles := filesParameters(tt.parameters)
			assert.Equal(t, tt.hasFiles, hasFiles)
			assert.Equal(t, tt.files, files)
		})
	}
}

type paramsWithFile struct {
	N int `json:"n"`
}

func (p *paramsWithFile) fileParameters() map[string]ta.NamedReader {
	return map[string]ta.NamedReader{
		"test": &testNamedReader{},
	}
}

type notStructParamsWithFile string

func (p *notStructParamsWithFile) fileParameters() map[string]ta.NamedReader {
	return map[string]ta.NamedReader{
		"test": &testNamedReader{},
	}
}

func TestBot_constructAndCallRequest(t *testing.T) {
	m := newMockedBot()

	params := struct {
		N int `json:"n"`
	}{
		N: 1,
	}

	url := m.Bot.apiURL + botPathPrefix + m.Bot.token + "/" + methodName

	expectedResp := &ta.Response{
		Ok: true,
	}

	paramsBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	expectedData := &ta.RequestData{
		ContentType: ta.ContentTypeJSON,
		BodyRaw:     paramsBytes,
	}

	t.Run("success_json", func(t *testing.T) {
		m.MockRequestConstructor.On("JSONRequest", params).
			Return(expectedData, nil).
			Times(1)

		m.MockAPICaller.On("Call", t.Context(), url, expectedData).
			Return(expectedResp, nil).
			Times(1)

		resp, err := m.Bot.constructAndCallRequest(t.Context(), methodName, params)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("error_json", func(t *testing.T) {
		m.MockRequestConstructor.On("JSONRequest", params).
			Return(nil, errTest).
			Times(1)

		resp, err := m.Bot.constructAndCallRequest(t.Context(), methodName, params)
		assert.ErrorIs(t, err, errTest)
		assert.Nil(t, resp)
	})

	t.Run("success_multipart", func(t *testing.T) {
		paramsFile := &paramsWithFile{N: 1}
		paramsMap := map[string]string{
			"n": "1",
		}

		paramsBytesFile, err := json.Marshal(paramsFile)
		assert.NoError(t, err)

		expectedDataFile := &ta.RequestData{
			ContentType: ta.ContentTypeJSON,
			BodyRaw:     paramsBytesFile,
		}

		m.MockRequestConstructor.On("MultipartRequest", paramsMap, mock.Anything).
			Return(expectedDataFile, nil).
			Times(1)

		m.MockAPICaller.On("Call", t.Context(), url, expectedDataFile).
			Return(expectedResp, nil).
			Times(1)

		resp, err := m.Bot.constructAndCallRequest(t.Context(), methodName, paramsFile)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("error_multipart", func(t *testing.T) {
		paramsFile := &paramsWithFile{N: 1}
		paramsMap := map[string]string{
			"n": "1",
		}

		m.MockRequestConstructor.On("MultipartRequest", paramsMap, mock.Anything).
			Return(nil, errTest).
			Times(1)

		resp, err := m.Bot.constructAndCallRequest(t.Context(), methodName, paramsFile)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("error_multipart_params", func(t *testing.T) {
		notStruct := notStructParamsWithFile("test")

		resp, err := m.Bot.constructAndCallRequest(t.Context(), methodName, &notStruct)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("error_call", func(t *testing.T) {
		m.MockRequestConstructor.On("JSONRequest", params).
			Return(expectedData, nil).
			Times(1)

		m.MockAPICaller.On("Call", t.Context(), url, expectedData).
			Return(nil, errTest).
			Times(1)

		resp, err := m.Bot.constructAndCallRequest(t.Context(), methodName, params)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestBot_performRequest(t *testing.T) {
	m := newMockedBot()

	params := struct {
		N int `json:"n"`
	}{
		N: 1,
	}

	t.Run("success", func(t *testing.T) {
		var result int

		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(&ta.RequestData{}, nil).
			Times(1)

		m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return(&ta.Response{
				Ok:     true,
				Result: bytes.NewBufferString("1").Bytes(),
				Error:  nil,
			}, nil).Once()

		err := m.Bot.performRequest(t.Context(), methodName, params, &result)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("success_unmarshal_second", func(t *testing.T) {
		var result1 int
		var result2 bool

		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(&ta.RequestData{}, nil).
			Times(1)

		m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return(&ta.Response{
				Ok:     true,
				Result: bytes.NewBufferString("true").Bytes(),
				Error:  nil,
			}, nil).Once()

		err := m.Bot.performRequest(t.Context(), methodName, params, &result1, &result2)
		assert.NoError(t, err)
		assert.Equal(t, 0, result1)
		assert.True(t, result2)
	})

	t.Run("error_not_ok", func(t *testing.T) {
		var result int

		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(&ta.RequestData{}, nil).
			Times(1)

		m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return(&ta.Response{
				Ok:     false,
				Result: nil,
				Error:  &ta.Error{},
			}, nil).Once()

		err := m.Bot.performRequest(t.Context(), methodName, params, &result)
		assert.Error(t, err)
	})

	t.Run("error_construct_and_call", func(t *testing.T) {
		var result int

		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(nil, errTest).
			Times(1)

		err := m.Bot.performRequest(t.Context(), methodName, params, &result)
		assert.Error(t, err)
	})

	t.Run("error_unmarshal", func(t *testing.T) {
		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(&ta.RequestData{}, nil).
			Times(1)

		m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return(&ta.Response{
				Ok:     true,
				Result: bytes.NewBufferString("1").Bytes(),
				Error:  nil,
			}, nil).Once()

		var stringResult string
		err := m.Bot.performRequest(t.Context(), methodName, params, &stringResult)
		assert.Error(t, err)
		assert.Empty(t, stringResult)
	})

	t.Run("error_warning", func(t *testing.T) {
		var result int

		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(&ta.RequestData{}, nil).
			Times(1)

		m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return(&ta.Response{
				Ok:     true,
				Result: bytes.NewBufferString("1").Bytes(),
				Error:  &ta.Error{ErrorCode: 1},
			}, nil).Once()

		err := m.Bot.performRequest(t.Context(), methodName, params, &result)
		assert.Equal(t, &ta.Error{ErrorCode: 1}, err)
		assert.Equal(t, 1, result)
	})
}

func Test_isNil(t *testing.T) {
	var n *int
	a := 1
	m := &a

	tests := []struct {
		name  string
		i     any
		isNil bool
	}{
		{
			name:  "nil",
			i:     nil,
			isNil: true,
		},
		{
			name:  "nil_ptr",
			i:     n,
			isNil: true,
		},
		{
			name:  "value",
			i:     m,
			isNil: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isNil, isNil(tt.i))
		})
	}
}

func TestToPtr(t *testing.T) {
	assert.True(t, *ToPtr(true))
	assert.False(t, *ToPtr(false))

	assert.Empty(t, *ToPtr(""))
	assert.Equal(t, "a", *ToPtr("a"))
}

func TestBot_ID_and_Username(t *testing.T) {
	m := newMockedBot()

	m.MockRequestConstructor.On("JSONRequest", nil).
		Return(&ta.RequestData{}, nil)
	m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).
		Return(telegoResponse(t, &User{
			ID:       123,
			Username: "test",
		}), nil)

	id := m.Bot.ID()
	assert.Equal(t, int64(123), id)

	username := m.Bot.Username()
	assert.Equal(t, "test", username)

	id = m.Bot.ID()
	assert.Equal(t, int64(123), id)
}

func Test_logRequestWithFiles(t *testing.T) {
	debug := &strings.Builder{}
	parameters := map[string]string{
		"foo": "bar",
	}
	files := map[string]ta.NamedReader{
		"file1":                  testNamedReader{},
		testNamedReader{}.Name(): testNamedReader{},
		"fileNil":                nil,
	}

	logRequestWithFiles(debug, parameters, files)
	assert.Equal(t, `parameters: {"foo":"bar"}, files: {"file1": "test", "test"}`, debug.String())
}

func Test_logRequest(t *testing.T) {
	debug := &strings.Builder{}
	parameters := map[string]string{
		"foo": "bar",
	}

	logRequest(debug, parameters)
	assert.Equal(t, `parameters: {"foo":"bar"}`, debug.String())
}
