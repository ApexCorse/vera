package codegen

import (
	"fmt"
	"io"
	"strings"

	"github.com/ApexCorse/vera"
)

func GenerateHeader(w io.Writer, config *vera.Config) error {
	s := fmt.Sprintf(includeFile, generateEncodingFunctionsDeclarations(config))
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}
	return nil
}

func GenerateSource(w io.Writer, config *vera.Config, headerFile string) error {
	var sb strings.Builder

	if headerFile != "" {
		sb.WriteString("#include \"" + headerFile + "\"\n")
	}

	sb.WriteString(sourceFileIncludes + "\n\n")
	sb.WriteString(utilFunctions + "\n\n")
	sb.WriteString(decodeSignalFunc + "\n\n")
	sb.WriteString(decodeMessageFunc + "\n\n")

	sb.WriteString(`vera_err_t vera_decode_can_frame(
	vera_can_rx_frame_t*  	frame,
	vera_decoding_result_t* result
) {
	switch (frame->id) {`)

	for _, m := range config.Messages {
		sb.WriteString(fmt.Sprintf("\n\t\tcase %#x: {", m.ID))
		sb.WriteString(fmt.Sprintf(`
			vera_message_t message = {
				.id = %#x,
				.name = "%s",
				.dlc = %d,
				.n_signals = %d
			};`,
			m.ID,
			m.Name,
			m.Length,
			len(m.Signals),
		))

		sb.WriteString(fmt.Sprintf(`
			message.signals = (vera_signal_t*)malloc(sizeof(vera_signal_t)*%d);`, len(m.Signals)))
		for i, s := range m.Signals {
			sb.WriteString(fmt.Sprintf(`
			message.signals[%d] = (vera_signal_t){
				.name = "%s",
				.unit = "%s",
				.start_bit = %d,
				.dlc = %d,
				.endianness = %d,
				.sign = %t,
				.integer_figures = %d,
				.decimal_figures = %d,
				.factor = %.4f,
				.offset = %.4f,
				.min = %.4f,
				.max = %.4f,
				.topic = "%s"
			};`,
				i,
				s.Name,
				s.Unit,
				s.StartBit,
				s.Length,
				s.Endianness,
				s.Signed,
				s.IntegerFigures,
				s.DecimalFigures,
				s.Factor,
				s.Offset,
				s.Min,
				s.Max,
				s.Topic,
			))
		}

		sb.WriteString(`
			vera_err_t err = _decode_message(
				frame,
				&message,
				result
			);
			if (err != vera_err_ok) {
				free(message.signals);
				return err;
			}
			break;
		}`)
	}
	sb.WriteString("\n\t}")
	sb.WriteString("\n\treturn vera_err_ok;")

	sb.WriteString("\n}\n")

	encodingFunctionsDefinitions := generateEncodingFunctionsDefinitions(config)
	if encodingFunctionsDefinitions != "" {
		sb.WriteString("\n")
		sb.WriteString(encodingFunctionsDefinitions)
	}

	if _, err := w.Write([]byte(sb.String())); err != nil {
		return err
	}
	return nil
}

func generateEncodingFunctionsDeclarations(config *vera.Config) string {
	s := ""
	for i, message := range config.Messages {
		s += "vera_err_t vera_encode_" + message.Name + "(\n"
		s += "\tvera_can_tx_frame_t* frame,\n"
		for j, signal := range message.Signals {
			s += "\tuint64_t " + signal.Name
			if j < len(message.Signals)-1 {
				s += ","
			}
			s += "\n"
		}

		s += ");"
		if i < len(config.Messages)-1 {
			s += "\n\n"
		}
	}

	return s
}

func generateEncodingFunctionsDefinitions(config *vera.Config) string {
	s := ""
	for i, message := range config.Messages {
		s += "vera_err_t vera_encode_" + message.Name + "(\n"
		s += "\tvera_can_tx_frame_t* frame,\n"
		for j, signal := range message.Signals {
			s += "\tuint64_t " + signal.Name
			if j < len(message.Signals)-1 {
				s += ","
			}
			s += "\n"
		}

		s += ") {\n"
		s += "\tif (!frame) return vera_err_null_arg;\n\n"

		s += "\tmemset(frame->data, 0, sizeof(uint8_t)*8);\n"
		s += "\tframe->id = " + fmt.Sprintf("0x%X", message.ID) + ";\n"
		s += "\tframe->dlc = " + fmt.Sprintf("%d", message.Length) + ";\n"

		if len(message.Signals) > 0 {
			s += "\n"
		}
		for _, signal := range message.Signals {
			s += fmt.Sprintf(
				"\t_insert_data_in_payload(frame->data, %s, %d, %d);\n",
				signal.Name,
				signal.StartBit,
				signal.Length,
			)
		}

		s += "\treturn vera_err_ok;\n"
		s += "}"
		if i < len(config.Messages)-1 {
			s += "\n\n"
		}
	}

	return s
}
