package codegen

import (
	"fmt"
	"io"
	"strings"

	"github.com/ApexCorse/vera/internal/parser"
)

func GenerateHeader(w io.Writer, config *parser.Config) error {
	if _, err := w.Write([]byte(includeFile)); err != nil {
		return err
	}
	return nil
}

func GenerateSource(w io.Writer, config *parser.Config, headerFile string) error {
	var sb strings.Builder

	if headerFile != "" {
		sb.WriteString("#include \"" + headerFile + "\"\n")
	}

	sb.WriteString(sourceFileIncludes + "\n\n")
	sb.WriteString(decodeSignalFunc + "\n\n")
	sb.WriteString(decodeMessageFunc + "\n\n")

	sb.WriteString(`vera_err_t vera_decode_can_frame(
	vera_can_rx_frame_t* frame,
	vera_decoded_signal_t* decoded_signals
) {
	switch (frame->id) {`)

	for _, m := range config.Messages {
		sb.WriteString(fmt.Sprintf("\n\t\tcase %#x:", m.ID))
		sb.WriteString("\n\t\t\tbreak;")
	}
	sb.WriteString("\n\t}")
	sb.WriteString("\n\treturn vera_err_ok;")

	sb.WriteString("\n}")

	if _, err := w.Write([]byte(sb.String())); err != nil {
		return err
	}
	return nil
}
