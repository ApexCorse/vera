package stm32hal

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ApexCorse/vera"
)

func TestGenerateSourcePassesFramePointerToCoreDecoder(t *testing.T) {
	var source bytes.Buffer
	if err := GenerateSource(&source, &vera.Config{}); err != nil {
		t.Fatalf("GenerateSource() error = %v", err)
	}

	if !strings.Contains(source.String(), "vera_decode_can_frame(&vera_frame, result)") {
		t.Error("generated STM32 HAL decoder must pass the address of its local vera frame")
	}
}
