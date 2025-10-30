package vera

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceNewLineCharacters(t *testing.T) {
	t.Run("should replace Windows line endings with Unix line endings", func(t *testing.T) {
		a := assert.New(t)

		input := "line1\r\nline2\r\nline3"
		expected := "line1\nline2\nline3"

		result := replaceNewLineCharacters(input)
		a.Equal(expected, result)
	})

	t.Run("should leave Unix line endings unchanged", func(t *testing.T) {
		a := assert.New(t)

		input := "line1\nline2\nline3"
		expected := "line1\nline2\nline3"

		result := replaceNewLineCharacters(input)
		a.Equal(expected, result)
	})

	t.Run("should handle empty string", func(t *testing.T) {
		a := assert.New(t)

		input := ""
		expected := ""

		result := replaceNewLineCharacters(input)
		a.Equal(expected, result)
	})

	t.Run("should handle string with no line endings", func(t *testing.T) {
		a := assert.New(t)

		input := "single line"
		expected := "single line"

		result := replaceNewLineCharacters(input)
		a.Equal(expected, result)
	})

	t.Run("should handle mixed line endings", func(t *testing.T) {
		a := assert.New(t)

		input := "line1\r\nline2\nline3\r\nline4"
		expected := "line1\nline2\nline3\nline4"

		result := replaceNewLineCharacters(input)
		a.Equal(expected, result)
	})
}
