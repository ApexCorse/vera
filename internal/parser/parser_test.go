package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("should return config struct with 2 messages", func(t *testing.T) {
		a := assert.New(t)

		configStr := `BO_ 123 EngineSpeed: 24 Engine
	SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
	SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway
BO_ 123 EngineSpeed: 24 Engine
	SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
	SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway`
		reader := strings.NewReader(configStr)

		config, err := Parse(reader)
		a.Nil(err)
		a.NotNil(config)
		a.Len(config.Messages, 2)
	})
}

func TestParseMessageInstruction(t *testing.T) {
	t.Run("should return message struct on correct definition", func(t *testing.T) {
		a := assert.New(t)

		messageStr := `BO_ 123 EngineSpeed: 24 Engine
	SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
	SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway`
		message, err := parseMessageInstruction(messageStr)

		a.Nil(err)
		a.NotNil(message)
		a.Equal(uint32(123), message.ID)
		a.Equal("EngineSpeed", message.Name)
		a.Equal(uint8(24), message.Length)
		a.Equal(Node("Engine"), message.Transmitter)

		a.Len(message.Signals, 2)
		a.Equal("EngineSpeed", message.Signals[0].Name)
		a.Equal("OilTemperature", message.Signals[1].Name)
		a.Equal(uint8(0), message.Signals[0].StartByte)
		a.Equal(uint8(16), message.Signals[1].StartByte)
		a.Equal(uint8(16), message.Signals[0].Length)
		a.Equal(uint8(8), message.Signals[1].Length)
		a.Equal(BigEndian, message.Signals[0].Endianness)
		a.Equal(BigEndian, message.Signals[1].Endianness)
		a.False(message.Signals[0].Signed)
		a.True(message.Signals[1].Signed)
		a.Equal(float32(0.1), message.Signals[0].Factor)
		a.Equal(float32(1), message.Signals[1].Factor)
		a.Equal(float32(0), message.Signals[0].Offset)
		a.Equal(float32(-40), message.Signals[1].Offset)
		a.Equal(float32(0), message.Signals[0].Min)
		a.Equal(float32(-40), message.Signals[1].Min)
		a.Equal(float32(8000), message.Signals[0].Max)
		a.Equal(float32(150), message.Signals[1].Max)
		a.Equal("RPM", message.Signals[0].Unit)
		a.Equal("ºC", message.Signals[1].Unit)
		a.Len(message.Signals[0].Receivers, 1)
		a.Len(message.Signals[1].Receivers, 2)
		a.Equal([]Node{"DriverGateway"}, message.Signals[0].Receivers)
		a.Equal([]Node{"DriverGateway", "EngineGateway"}, message.Signals[1].Receivers)
	})
}

func TestParseMessageDefinition(t *testing.T) {
	t.Run("should return message struct on correct definition", func(t *testing.T) {
		a := assert.New(t)

		messageStr := `BO_ 123 EngineSpeed: 24 Engine`
		message, err := parseMessageDefinition(messageStr)

		a.Nil(err)
		a.NotNil(message)
		a.Equal(uint32(123), message.ID)
		a.Equal("EngineSpeed", message.Name)
		a.Equal(uint8(24), message.Length)
		a.Equal(Node("Engine"), message.Transmitter)
	})
}

func TestParseMessageSignals(t *testing.T) {
	t.Run("should return signal struct array on correct input", func(t *testing.T) {
		a := assert.New(t)

		signalsStr := `	SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
	SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway`
		signals, err := parseMessageSignals(signalsStr)

		a.Nil(err)
		a.NotNil(signals)
		a.Len(signals, 2)

		a.Equal("EngineSpeed", signals[0].Name)
		a.Equal("OilTemperature", signals[1].Name)
		a.Equal(uint8(0), signals[0].StartByte)
		a.Equal(uint8(16), signals[1].StartByte)
		a.Equal(uint8(16), signals[0].Length)
		a.Equal(uint8(8), signals[1].Length)
		a.Equal(BigEndian, signals[0].Endianness)
		a.Equal(BigEndian, signals[1].Endianness)
		a.False(signals[0].Signed)
		a.True(signals[1].Signed)
		a.Equal(float32(0.1), signals[0].Factor)
		a.Equal(float32(1), signals[1].Factor)
		a.Equal(float32(0), signals[0].Offset)
		a.Equal(float32(-40), signals[1].Offset)
		a.Equal(float32(0), signals[0].Min)
		a.Equal(float32(-40), signals[1].Min)
		a.Equal(float32(8000), signals[0].Max)
		a.Equal(float32(150), signals[1].Max)
		a.Equal("RPM", signals[0].Unit)
		a.Equal("ºC", signals[1].Unit)
		a.Len(signals[0].Receivers, 1)
		a.Len(signals[1].Receivers, 2)
		a.Equal(signals[0].Receivers, []Node{"DriverGateway"})
		a.Equal(signals[1].Receivers, []Node{"DriverGateway", "EngineGateway"})
	})

	t.Run("should return signal struct array on correct input with additional spaces", func(t *testing.T) {
		a := assert.New(t)

		signalsStr := `	SG_ EngineSpeed :    0|16@1+ (0.1,0)  [0|8000]     "RPM" DriverGateway
	SG_   OilTemperature :    16|8@1-   (1,-40) [-40|150]  "ºC"   DriverGateway,EngineGateway`
		signals, err := parseMessageSignals(signalsStr)

		a.Nil(err)
		a.NotNil(signals)
		a.Len(signals, 2)

		a.Equal("EngineSpeed", signals[0].Name)
		a.Equal("OilTemperature", signals[1].Name)
		a.Equal(uint8(0), signals[0].StartByte)
		a.Equal(uint8(16), signals[1].StartByte)
		a.Equal(uint8(16), signals[0].Length)
		a.Equal(uint8(8), signals[1].Length)
		a.Equal(BigEndian, signals[0].Endianness)
		a.Equal(BigEndian, signals[1].Endianness)
		a.False(signals[0].Signed)
		a.True(signals[1].Signed)
		a.Equal(float32(0.1), signals[0].Factor)
		a.Equal(float32(1), signals[1].Factor)
		a.Equal(float32(0), signals[0].Offset)
		a.Equal(float32(-40), signals[1].Offset)
		a.Equal(float32(0), signals[0].Min)
		a.Equal(float32(-40), signals[1].Min)
		a.Equal(float32(8000), signals[0].Max)
		a.Equal(float32(150), signals[1].Max)
		a.Equal("RPM", signals[0].Unit)
		a.Equal("ºC", signals[1].Unit)
		a.Len(signals[0].Receivers, 1)
		a.Len(signals[1].Receivers, 2)
		a.Equal([]Node{"DriverGateway"}, signals[0].Receivers)
		a.Equal([]Node{"DriverGateway", "EngineGateway"}, signals[1].Receivers)
	})

	t.Run("fails because it lacks tabs", func(t *testing.T) {
		a := assert.New(t)

		// Missing tab
		signalsStr := `SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
	SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway`
		signals, err := parseMessageSignals(signalsStr)

		a.Error(err)
		a.Nil(signals)
	})
}
