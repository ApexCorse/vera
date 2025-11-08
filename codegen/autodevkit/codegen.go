package autodevkit

import (
	"io"
)

func GenerateHeader(w io.Writer) error {
	if _, err := w.Write([]byte(includeFile)); err != nil {
		return err
	}
	return nil
}

func GenerateSource(w io.Writer) error {
	if _, err := w.Write([]byte(sourceFile)); err != nil {
		return err
	}
	return nil
}
