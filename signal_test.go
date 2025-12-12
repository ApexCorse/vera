package vera

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSignalBitInfo(t *testing.T) {
	t.Run("should parse bit info with little endian", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		message := &Message{}
		err := signal.parseBitInfo(message, "0|16@0+")
		a.Nil(err)
		a.Equal(uint8(0), signal.StartBit)
		a.Equal(uint8(16), signal.Length)
		a.Equal(LittleEndian, signal.Endianness)
		a.False(signal.Signed)
	})

	t.Run("should parse bit info with signed signal", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		message := &Message{}
		err := signal.parseBitInfo(message, "0|16@1-")
		a.Nil(err)
		a.Equal(uint8(0), signal.StartBit)
		a.Equal(uint8(16), signal.Length)
		a.Equal(BigEndian, signal.Endianness)
		a.True(signal.Signed)
	})

	t.Run("should return error for invalid @ separator", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		message := &Message{}
		err := signal.parseBitInfo(message, "0|16")
		a.Error(err)
		a.Contains(err.Error(), "invalid bit info")
	})

	t.Run("should return error for invalid pipe separator", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		message := &Message{}
		err := signal.parseBitInfo(message, "016@1+")
		a.Error(err)
		a.Contains(err.Error(), "invalid start bit and length")
	})

	t.Run("should return error for invalid start bit", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		message := &Message{}
		err := signal.parseBitInfo(message, "abc|16@1+")
		a.Error(err)
		a.Contains(err.Error(), "invalid start bit")
	})

	t.Run("should return error for invalid length", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		message := &Message{}
		err := signal.parseBitInfo(message, "0|abc@1+")
		a.Error(err)
		a.Contains(err.Error(), "invalid length")
	})

	t.Run("should return error for invalid endianness", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		message := &Message{}
		err := signal.parseBitInfo(message, "0|16@x+")
		a.Error(err)
		a.Contains(err.Error(), "invalid bit order")
	})

	t.Run("should return error for invalid signed character", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		message := &Message{}
		err := signal.parseBitInfo(message, "0|16@1x")
		a.Error(err)
		a.Contains(err.Error(), "invalid signed")
	})

	t.Run("should return error for too short other info", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		message := &Message{}
		err := signal.parseBitInfo(message, "0|16@1")
		a.Error(err)
		a.Contains(err.Error(), "invalid bit order and signed")
	})

}

func TestParseSignalFactorOffset(t *testing.T) {
	t.Run("should parse factor and offset", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseFactorOffset("(0.5,-10)")
		a.Nil(err)
		a.Equal(float32(0.5), signal.Factor)
		a.Equal(float32(-10), signal.Offset)
	})

	t.Run("should return error for missing parentheses", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseFactorOffset("0.5,-10")
		a.Error(err)
		a.Contains(err.Error(), "invalid factor and offset")
	})

	t.Run("should return error for invalid structure", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseFactorOffset("(0.5)")
		a.Error(err)
		a.Contains(err.Error(), "invalid factor and offset")
	})

	t.Run("should return error for invalid factor", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseFactorOffset("(abc,-10)")
		a.Error(err)
		a.Contains(err.Error(), "invalid factor")
	})

	t.Run("should return error for invalid offset", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseFactorOffset("(0.5,xyz)")
		a.Error(err)
		a.Contains(err.Error(), "invalid offset")
	})
}

func TestParseSignalMinMax(t *testing.T) {
	t.Run("should parse min and max", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseMinMax("[-40|150]")
		a.Nil(err)
		a.Equal(float32(-40), signal.Min)
		a.Equal(float32(150), signal.Max)
	})

	t.Run("should return error for missing brackets", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseMinMax("-40|150")
		a.Error(err)
		a.Contains(err.Error(), "invalid min/max")
	})

	t.Run("should return error for invalid structure", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseMinMax("[-40]")
		a.Error(err)
		a.Contains(err.Error(), "invalid min/max")
	})

	t.Run("should return error for invalid min", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseMinMax("[abc|150]")
		a.Error(err)
		a.Contains(err.Error(), "invalid min")
	})

	t.Run("should return error for invalid max", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseMinMax("[-40|xyz]")
		a.Error(err)
		a.Contains(err.Error(), "invalid max")
	})
}

func TestParseSignalUnit(t *testing.T) {
	t.Run("should parse unit", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseUnit("\"km/h\"")
		a.Nil(err)
		a.Equal("km/h", signal.Unit)
	})

	t.Run("should return error for missing quotes", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseUnit("km/h")
		a.Error(err)
		a.Contains(err.Error(), "invalid unit")
	})

	t.Run("should handle empty unit", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		err := signal.parseUnit("\"\"")
		a.Nil(err)
		a.Equal("", signal.Unit)
	})
}

func TestParseSignalReceivers(t *testing.T) {
	t.Run("should parse single receiver", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		signal.parseReceivers("Gateway")
		a.Len(signal.Receivers, 1)
		a.Equal(Node("Gateway"), signal.Receivers[0])
	})

	t.Run("should parse multiple receivers", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{}
		signal.parseReceivers("Gateway,Engine,Driver")
		a.Len(signal.Receivers, 3)
		a.Equal(Node("Gateway"), signal.Receivers[0])
		a.Equal(Node("Engine"), signal.Receivers[1])
		a.Equal(Node("Driver"), signal.Receivers[2])
	})
}

func TestSignalValidate(t *testing.T) {
	t.Run("should validate signal with valid values", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{
			Name:     "Speed",
			StartBit: 0,
			Length:   2,
			Factor:   0.1,
			Min:      0,
			Max:      100,
		}

		err := signal.Validate()
		a.Nil(err)
	})

	t.Run("should return error when start bit >= 64", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{
			Name:     "Speed",
			StartBit: 64,
			Length:   2,
			Factor:   0.1,
		}

		err := signal.Validate()
		a.Error(err)
	})

	t.Run("should return error when length > 64", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{
			Name:     "Speed",
			StartBit: 0,
			Length:   65,
			Factor:   0.1,
		}

		err := signal.Validate()
		a.Error(err)
	})

	t.Run("should return error when factor is zero", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{
			Name:     "Speed",
			StartBit: 0,
			Length:   2,
			Factor:   0,
		}

		err := signal.Validate()
		a.Error(err)
	})
}
