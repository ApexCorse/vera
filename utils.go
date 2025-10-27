package vera

import "strings"

func replaceNewLineCharacters(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
