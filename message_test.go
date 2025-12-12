package vera

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMessageInstruction(t *testing.T) {
	t.Run("should return message struct on correct definition", func(t *testing.T) {
		a := assert.New(t)

		messageStr := `BO_ 123 EngineSpeed: 3 Engine
SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway`
		lines := strings.Split(messageStr, "\n")
		message, err := NewMessageFromLines(lines, 0)

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

func TestParseMessageID(t *testing.T) {
	t.Run("should parse decimal message ID", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{}
		err := message.parseID("123")
		a.Nil(err)
		a.Equal(uint32(123), message.ID)
	})

	t.Run("should parse hexadecimal message ID", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{}
		err := message.parseID("0x7B")
		a.Nil(err)
		a.Equal(uint32(123), message.ID)
	})

	t.Run("should return error for invalid decimal ID", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{}
		err := message.parseID("abc")
		a.Error(err)
		a.Equal(uint32(0), message.ID)
	})

	t.Run("should return error for invalid hexadecimal ID", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{}
		err := message.parseID("0xGHI")
		a.Error(err)
		a.Equal(uint32(0), message.ID)
	})
}

func TestMessageValidate(t *testing.T) {
	t.Run("should validate message with valid DLC", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{
			ID:          123,
			Name:        "EngineSpeed",
			DLC:         1,
			Transmitter: "Engine",
			Signals: []Signal{
				{
					Name:     "Speed",
					StartBit: 0,
					Length:   2,
					Factor:   0.1,
					Min:      0,
					Max:      100,
				},
			},
		}

		err := message.Validate()
		a.Nil(err)
	})

	t.Run("should return error when message DLC > 8", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{
			ID:          123,
			Name:        "EngineSpeed",
			DLC:         9,
			Transmitter: "Engine",
		}

		err := message.Validate()
		a.Error(err)
	})

	t.Run("should return error when signal lengths exceed message DLC", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{
			ID:                 123,
			Name:               "EngineSpeed",
			DLC:                1,
			Transmitter:        "Engine",
			signalsTotalLength: 10,
		}

		err := message.Validate()
		a.Error(err)
	})

	t.Run("should return error for invalid signal", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{
			ID:          123,
			Name:        "EngineSpeed",
			DLC:         1,
			Transmitter: "Engine",
			Signals: []Signal{
				{
					Name:     "Speed",
					StartBit: 0,
					Length:   2,
					Factor:   0, // Invalid: factor is zero
				},
			},
		}

		fmt.Println(message.Signals[0].Factor)
		err := message.Validate()
		a.NotNil(err)
		a.Contains(err.Error(), "signal factor cannot be zero")
	})
}
