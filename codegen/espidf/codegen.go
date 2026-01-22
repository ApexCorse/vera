package espidf

import (
	"fmt"
	"io"

	"github.com/ApexCorse/vera"
)

func GenerateHeader(w io.Writer, config *vera.Config) error {
	s := fmt.Sprintf(headerFile, generateEncodingFunctionsDeclarations(config))
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}
	return nil
}

func GenerateSource(w io.Writer, config *vera.Config) error {
	s := fmt.Sprintf(sourceFile, generateEncodingFunctionsDefinitions(config))
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}
	return nil
}

func generateEncodingFunctionsDeclarations(config *vera.Config) string {
	s := ""
	for i, message := range config.Messages {
		s += "vera_err_t vera_encode_espidf_" + message.Name + "(\n"
		s += "\ttwai_frame_t* frame,\n"
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
		s += "vera_err_t vera_encode_espidf_" + message.Name + "(\n"
		s += "\ttwai_frame_t* frame,\n"
		for j, signal := range message.Signals {
			s += "\tuint64_t " + signal.Name
			if j < len(message.Signals)-1 {
				s += ","
			}
			s += "\n"
		}

		s += ") {\n"
		s += "\tif (!frame || !frame->buffer) return vera_err_null_arg;\n\n"
		s += "\tmemset(frame->buffer, 0, sizeof(uint8_t)*8);\n"
		s += "\tframe->header.id = " + fmt.Sprintf("0x%X", message.ID) + ";\n"
		s += "\tframe->header.dlc = " + fmt.Sprintf("%d", message.DLC) + ";\n"

		if len(message.Signals) > 0 {
			s += "\n"
		}
		for _, signal := range message.Signals {
			s += fmt.Sprintf(
				"\t_insert_data_in_payload(frame->buffer, %s, %d, %d);\n",
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
