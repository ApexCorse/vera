package vera

import (
	"fmt"
)

func errorAtLine(lineNumber int, format string, a ...any) error {
	errStr := fmt.Sprintf(format, a...)
	return fmt.Errorf("line %d: %s", lineNumber, errStr)
}
