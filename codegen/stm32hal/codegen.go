package stm32hal

import (
	"embed"
	"io"
	"text/template"

	"github.com/ApexCorse/vera"
)

//go:embed *.tmpl
var templateFiles embed.FS

func GenerateHeader(w io.Writer, config *vera.Config) error {
	headerTemplateContent, err := templateFiles.ReadFile("vera_stm32hal.h.tmpl")
	if err != nil {
		return err
	}

	headerTmpl, err := template.New("vera_stm32hal.h").Parse(string(headerTemplateContent))
	if err != nil {
		return err
	}

	if err := headerTmpl.Execute(w, config); err != nil {
		return nil
	}

	return nil
}

func GenerateSource(w io.Writer, config *vera.Config) error {
	sourceTemplateContent, err := templateFiles.ReadFile("vera_stm32hal.c.tmpl")
	if err != nil {
		return err
	}

	sourceTmpl, err := template.New("vera_stm32hal.c").Parse(string(sourceTemplateContent))
	if err != nil {
		return err
	}

	if err := sourceTmpl.Execute(w, config); err != nil {
		return nil
	}

	return nil
}
