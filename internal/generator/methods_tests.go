package main

import (
	"fmt"
	"strings"
)

func generateMethodsTests(methods tgMethods) {
	methodsTestsFile := openFile(generatedMethodsTestsFilename)
	defer func() { _ = methodsTestsFile.Close() }()

	data := strings.Builder{}

	data.WriteString(`package telego

import (
	"testing"

	"gitlab.com/asvedr/testify/assert"
	"gitlab.com/asvedr/testify/mock"

	ta "github.com/mymmrac/telego/telegoapi"
)
`)

	for _, m := range methods {
		data.WriteString(fmt.Sprintf("func TestBot_%s(t *testing.T) {\n", m.nameTitle))
		data.WriteString(`	m := newMockedBot()

	t.Run("success", func(t *testing.T) {
		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(data, nil).Once()
`)

		respVar := "emptyResp"
		actualVar := returnTypeToVar(m.returnType, m.nameTitle)
		expectedVar := fmt.Sprintf("expected%s", firstToUpper(actualVar))
		if m.hasReturnValue() {
			expectedData := strings.Replace(m.returnType, "*", "&", 1) + "{}"

			// Special case
			if expectedVar != "expectedMessage" {
				data.WriteString(fmt.Sprintf("\n\t\t%s := %s", expectedVar, expectedData))
			}

			respVar = "resp"
			if m.returnType == "*string" {
				expectedVar = "&" + expectedVar
			}
			data.WriteString(fmt.Sprintf("\n\t\tresp := telegoResponse(t, %s)", expectedVar))
		}

		parameters := ""
		if len(m.parameters) > 0 {
			parameters = "nil"
		}

		data.WriteString(`
		m.MockAPICaller.On("Call", mock.Anything, mock.Anything, mock.Anything).
			Return(` + respVar + `, nil).Once()`)
		data.WriteString("\n\n")

		if m.hasReturnValue() {
			data.WriteString(fmt.Sprintf(`		%s, err := m.Bot.%s(t.Context(), %s)
		assert.NoError(t, err)
		assert.Equal(t, %s, %s)`, actualVar, m.nameTitle, parameters, expectedVar, actualVar))
		} else {
			data.WriteString(fmt.Sprintf(`		err := m.Bot.%s(t.Context(), %s)
		assert.NoError(t, err)`, m.nameTitle, parameters))
		}

		data.WriteString("\n\t})\n\n")

		data.WriteString(`	t.Run("error", func(t *testing.T) {
		m.MockRequestConstructor.On("JSONRequest", mock.Anything).
			Return(nil, errTest).Once()`)
		data.WriteString("\n\n")

		if m.hasReturnValue() {
			data.WriteString(fmt.Sprintf(`		%s, err := m.Bot.%s(t.Context(), %s)
		assert.Error(t, err)
		assert.Nil(t, %s)`, actualVar, m.nameTitle, parameters, actualVar))
		} else {
			data.WriteString(fmt.Sprintf(`		err := m.Bot.%s(t.Context(), %s)
		assert.Error(t, err)`, m.nameTitle, parameters))
		}

		data.WriteString("\n\t})\n}\n\n")
	}

	_, err := methodsTestsFile.WriteString(uppercaseWords(data.String()))
	exitOnErr(err)

	formatFile(methodsTestsFile.Name())
}
