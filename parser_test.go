package vera

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("should return config struct with 2 messages", func(t *testing.T) {
		a := assert.New(t)

		configStr := `BO_ 123 EngineSpeed: 3 Engine
SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
	SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway
BO_ 123 EngineSpeed: 3 Engine
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

		messageStr := `BO_ 123 EngineSpeed: 3 Engine
SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway`
		message, err := parseMessageInstruction(messageStr)

		a.Nil(err)
		a.NotNil(message)
		a.Equal(uint32(123), message.ID)
		a.Equal("EngineSpeed", message.Name)
		a.Equal(uint8(3), message.DLC)
		a.Equal(Node("Engine"), message.Transmitter)

		a.Len(message.Signals, 2)
		a.Equal("EngineSpeed", message.Signals[0].Name)
		a.Equal("OilTemperature", message.Signals[1].Name)
		a.Equal(uint8(0), message.Signals[0].StartBit)
		a.Equal(uint8(16), message.Signals[1].StartBit)
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

		messageStr := `BO_ 123 EngineSpeed: 3 Engine`
		message, err := parseMessageDefinition(messageStr)

		a.Nil(err)
		a.NotNil(message)
		a.Equal(uint32(123), message.ID)
		a.Equal("EngineSpeed", message.Name)
		a.Equal(uint8(3), message.DLC)
		a.Equal(Node("Engine"), message.Transmitter)
	})
}

func TestParseMessageSignals(t *testing.T) {
	t.Run("should return signal struct array on correct input", func(t *testing.T) {
		a := assert.New(t)

		signalsStr := `SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway`
		signals, err := parseMessageSignals(signalsStr)

		a.Nil(err)
		a.NotNil(signals)
		a.Len(signals, 2)

		a.Equal("EngineSpeed", signals[0].Name)
		a.Equal("OilTemperature", signals[1].Name)
		a.Equal(uint8(0), signals[0].StartBit)
		a.Equal(uint8(16), signals[1].StartBit)
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

		signalsStr := `SG_ EngineSpeed :    0|16@1+ (0.1,0)  [0|8000]     "RPM" DriverGateway
SG_   OilTemperature :    16|8@1-   (1,-40) [-40|150]  "ºC"   DriverGateway,EngineGateway`
		signals, err := parseMessageSignals(signalsStr)

		a.Nil(err)
		a.NotNil(signals)
		a.Len(signals, 2)

		a.Equal("EngineSpeed", signals[0].Name)
		a.Equal("OilTemperature", signals[1].Name)
		a.Equal(uint8(0), signals[0].StartBit)
		a.Equal(uint8(16), signals[1].StartBit)
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

	t.Run("should return signal struct array on correct input (including integer and decimal part)", func(t *testing.T) {
		a := assert.New(t)

		signalsStr := `SG_ EngineSpeed : 0|16@1+(4,4) (0.1,0) [0|8000] "RPM" DriverGateway`
		signals, err := parseMessageSignals(signalsStr)

		a.Nil(err)
		a.NotNil(signals)
		a.Len(signals, 1)

		a.Equal("EngineSpeed", signals[0].Name)
		a.Equal(uint8(0), signals[0].StartBit)
		a.Equal(uint8(16), signals[0].Length)
		a.Equal(BigEndian, signals[0].Endianness)
		a.False(signals[0].Signed)
		a.Equal(float32(0.1), signals[0].Factor)
		a.Equal(float32(0), signals[0].Offset)
		a.Equal(float32(0), signals[0].Min)
		a.Equal(float32(8000), signals[0].Max)
		a.Equal("RPM", signals[0].Unit)
		a.Len(signals[0].Receivers, 1)
		a.Equal([]Node{"DriverGateway"}, signals[0].Receivers)

		a.Equal(uint8(4), signals[0].IntegerFigures)
		a.Equal(uint8(4), signals[0].DecimalFigures)
	})
}

func TestParseMessageID(t *testing.T) {
	t.Run("should parse decimal message ID", func(t *testing.T) {
		a := assert.New(t)

		messageID, err := parseMessageID("123")
		a.Nil(err)
		a.Equal(uint32(123), messageID)
	})

	t.Run("should parse hexadecimal message ID", func(t *testing.T) {
		a := assert.New(t)

		messageID, err := parseMessageID("0x7B")
		a.Nil(err)
		a.Equal(uint32(123), messageID)
	})

	t.Run("should return error for invalid decimal ID", func(t *testing.T) {
		a := assert.New(t)

		messageID, err := parseMessageID("abc")
		a.Equal(ErrorMessageIDNotInteger, err)
		a.Equal(uint32(0), messageID)
	})

	t.Run("should return error for invalid hexadecimal ID", func(t *testing.T) {
		a := assert.New(t)

		messageID, err := parseMessageID("0xGHI")
		a.Equal(ErrorMessageIDNotInteger, err)
		a.Equal(uint32(0), messageID)
	})
}

func TestParseSignalTopic(t *testing.T) {
	t.Run("should parse valid signal topic", func(t *testing.T) {
		a := assert.New(t)

		topic, err := parseSignalTopic("TP_ EngineSpeed vehicle/engine/speed")
		a.Nil(err)
		a.NotNil(topic)
		a.Equal("EngineSpeed", topic.Signal)
		a.Equal("vehicle/engine/speed", topic.Topic)
	})

	t.Run("should return error for invalid structure", func(t *testing.T) {
		a := assert.New(t)

		topic, err := parseSignalTopic("TP_ EngineSpeed")
		a.Error(err)
		a.Nil(topic)
		a.Contains(err.Error(), "signal topic has wrong structure")
	})

	t.Run("should return error for too many fields", func(t *testing.T) {
		a := assert.New(t)

		topic, err := parseSignalTopic("TP_ EngineSpeed vehicle/engine/speed extra")
		a.Error(err)
		a.Nil(topic)
	})
}

func TestParse_WithTopics(t *testing.T) {
	t.Run("should parse config with topics", func(t *testing.T) {
		a := assert.New(t)

		configStr := `BO_ 123 EngineSpeed: 8 Engine
	SG_ Speed : 0|2@1+ (0.1,0) [0|100] "km/h" Gateway
TP_ Speed vehicle/engine/speed`
		reader := strings.NewReader(configStr)

		config, err := Parse(reader)
		a.Nil(err)
		a.NotNil(config)
		a.Len(config.Messages, 1)
		a.Len(config.Topics, 1)
		a.Equal("Speed", config.Topics[0].Signal)
		a.Equal("vehicle/engine/speed", config.Topics[0].Topic)
	})

	t.Run("should return empty config for empty input", func(t *testing.T) {
		a := assert.New(t)

		configStr := ``
		reader := strings.NewReader(configStr)

		config, err := Parse(reader)
		a.Nil(err)
		a.NotNil(config)
		a.Len(config.Messages, 0)
		a.Len(config.Topics, 0)
	})

	t.Run("should skip lines that are not BO_ or TP_", func(t *testing.T) {
		a := assert.New(t)

		configStr := `CM_ "This is a comment"
BO_ 123 EngineSpeed: 8 Engine
	SG_ Speed : 0|2@1+ (0.1,0) [0|100] "km/h" Gateway
BA_ "AttributeName" "AttributeValue"
TP_ Speed vehicle/engine/speed`
		reader := strings.NewReader(configStr)

		config, err := Parse(reader)
		a.Nil(err)
		a.NotNil(config)
		a.Len(config.Messages, 1)
		a.Len(config.Topics, 1)
	})
}

func TestParseMessageBitInfo(t *testing.T) {
	t.Run("should parse bit info with little endian", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16@0+")
		a.Nil(err)
		a.Equal(uint8(0), signal.StartBit)
		a.Equal(uint8(16), signal.Length)
		a.Equal(LittleEndian, signal.Endianness)
		a.False(signal.Signed)
	})

	t.Run("should parse bit info with signed signal", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16@1-")
		a.Nil(err)
		a.Equal(uint8(0), signal.StartBit)
		a.Equal(uint8(16), signal.Length)
		a.Equal(BigEndian, signal.Endianness)
		a.True(signal.Signed)
	})

	t.Run("should return error for invalid @ separator", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16")
		a.Error(err)
		a.Contains(err.Error(), "invalid bit info")
	})

	t.Run("should return error for invalid pipe separator", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "016@1+")
		a.Error(err)
		a.Contains(err.Error(), "invalid start bit and length")
	})

	t.Run("should return error for invalid start bit", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "abc|16@1+")
		a.Error(err)
		a.Contains(err.Error(), "invalid start bit")
	})

	t.Run("should return error for invalid length", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|abc@1+")
		a.Error(err)
		a.Contains(err.Error(), "invalid length")
	})

	t.Run("should return error for invalid endianness", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16@x+")
		a.Error(err)
		a.Contains(err.Error(), "invalid bit order")
	})

	t.Run("should return error for invalid signed character", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16@1x")
		a.Error(err)
		a.Contains(err.Error(), "invalid signed")
	})

	t.Run("should return error for too short other info", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16@1")
		a.Error(err)
		a.Contains(err.Error(), "invalid bit order and signed")
	})

	t.Run("should return error for invalid decimal format parenthesis", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16@1+[4,4]")
		a.Error(err)
		a.Contains(err.Error(), "invalid decimal format")
	})

	t.Run("should return error for invalid decimal format structure", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16@1+(4)")
		a.Error(err)
		a.Contains(err.Error(), "invalid decimal format")
	})

	t.Run("should return error for invalid integer figures", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16@1+(a,4)")
		a.Error(err)
		a.Contains(err.Error(), "invalid decimal format")
	})

	t.Run("should return error for invalid decimal figures", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageBitInfo(signal, 0, "0|16@1+(4,b)")
		a.Error(err)
		a.Contains(err.Error(), "invalid decimal format")
	})
}

func TestParseMessageFactorOffset(t *testing.T) {
	t.Run("should parse factor and offset", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageFactorOffset(signal, 0, "(0.5,-10)")
		a.Nil(err)
		a.Equal(float32(0.5), signal.Factor)
		a.Equal(float32(-10), signal.Offset)
	})

	t.Run("should return error for missing parentheses", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageFactorOffset(signal, 0, "0.5,-10")
		a.Error(err)
		a.Contains(err.Error(), "invalid factor and offset")
	})

	t.Run("should return error for invalid structure", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageFactorOffset(signal, 0, "(0.5)")
		a.Error(err)
		a.Contains(err.Error(), "invalid factor and offset")
	})

	t.Run("should return error for invalid factor", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageFactorOffset(signal, 0, "(abc,-10)")
		a.Error(err)
		a.Contains(err.Error(), "invalid factor")
	})

	t.Run("should return error for invalid offset", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageFactorOffset(signal, 0, "(0.5,xyz)")
		a.Error(err)
		a.Contains(err.Error(), "invalid offset")
	})
}

func TestParseMessageMinMax(t *testing.T) {
	t.Run("should parse min and max", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageMinMax(signal, 0, "[-40|150]")
		a.Nil(err)
		a.Equal(float32(-40), signal.Min)
		a.Equal(float32(150), signal.Max)
	})

	t.Run("should return error for missing brackets", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageMinMax(signal, 0, "-40|150")
		a.Error(err)
		a.Contains(err.Error(), "invalid min/max")
	})

	t.Run("should return error for invalid structure", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageMinMax(signal, 0, "[-40]")
		a.Error(err)
		a.Contains(err.Error(), "invalid min/max")
	})

	t.Run("should return error for invalid min", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageMinMax(signal, 0, "[abc|150]")
		a.Error(err)
		a.Contains(err.Error(), "invalid min")
	})

	t.Run("should return error for invalid max", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageMinMax(signal, 0, "[-40|xyz]")
		a.Error(err)
		a.Contains(err.Error(), "invalid max")
	})
}

func TestParseMessageUnit(t *testing.T) {
	t.Run("should parse unit", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageUnit(signal, 0, "\"km/h\"")
		a.Nil(err)
		a.Equal("km/h", signal.Unit)
	})

	t.Run("should return error for missing quotes", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageUnit(signal, 0, "km/h")
		a.Error(err)
		a.Contains(err.Error(), "invalid unit")
	})

	t.Run("should handle empty unit", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := parseMessageUnit(signal, 0, "\"\"")
		a.Nil(err)
		a.Equal("", signal.Unit)
	})
}

func TestParseMessageReceivers(t *testing.T) {
	t.Run("should parse single receiver", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		parseMessageReceivers(signal, "Gateway")
		a.Len(signal.Receivers, 1)
		a.Equal(Node("Gateway"), signal.Receivers[0])
	})

	t.Run("should parse multiple receivers", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		parseMessageReceivers(signal, "Gateway,Engine,Driver")
		a.Len(signal.Receivers, 3)
		a.Equal(Node("Gateway"), signal.Receivers[0])
		a.Equal(Node("Engine"), signal.Receivers[1])
		a.Equal(Node("Driver"), signal.Receivers[2])
	})
}
