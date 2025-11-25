package stm32hal

import (
	"fmt"
	"io"
	"math"

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
		s += "vera_err_t vera_encode_stm32hal_" + message.Name + "(\n"
		s += "\tCAN_TxHeaderTypeDef* frame,\n"
		s += "\tuint8_t* data,\n"
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
		s += "vera_err_t vera_encode_stm32hal_" + message.Name + "(\n"
		s += "\tCAN_TxHeaderTypeDef* frame,\n"
		s += "\tuint8_t* data,\n"
		for j, signal := range message.Signals {
			s += "\tuint64_t " + signal.Name
			if j < len(message.Signals)-1 {
				s += ","
			}
			s += "\n"
		}

		s += ") {\n"
		s += "\tif (!frame) return vera_err_null_arg;\n\n"

		s += "\tmemset(data, 0, sizeof(uint8_t)*8);\n"
		s += "\tframe->ID = " + fmt.Sprintf("0x%X", message.ID) + ";\n"
		s += "\tframe->DLC = " + fmt.Sprintf("%.0f", math.Ceil(float64(message.Length)/8)) + ";\n"

		if len(message.Signals) > 0 {
			s += "\n"
		}
		for _, signal := range message.Signals {
			s += fmt.Sprintf(
				"\t_insert_data_in_payload(data, %s, %d, %d);\n",
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
