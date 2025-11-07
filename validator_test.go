package vera

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	t.Run("should validate valid config", func(t *testing.T) {
		a := assert.New(t)

		config := &Config{
			Messages: []Message{
				{
					ID:          123,
					Name:        "EngineSpeed",
					Length:      8,
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
				},
			},
			Topics: []SignalTopic{
				{
					Topic:  "vehicle/engine/speed",
					Signal: "Speed",
				},
			},
		}

		err := config.Validate()
		a.Nil(err)
		a.Equal("vehicle/engine/speed", config.Messages[0].Signals[0].Topic)
	})

	t.Run("should return error for invalid topic", func(t *testing.T) {
		a := assert.New(t)

		config := &Config{
			Topics: []SignalTopic{
				{
					Topic:  "",
					Signal: "Speed",
				},
			},
		}

		err := config.Validate()
		a.NotNil(err)
		a.Contains(err.Error(), "topic Nº0")
	})

	t.Run("should return error for duplicate signal topic", func(t *testing.T) {
		a := assert.New(t)

		config := &Config{
			Topics: []SignalTopic{
				{
					Topic:  "vehicle/engine/speed",
					Signal: "Speed",
				},
				{
					Topic:  "vehicle/engine/rpm",
					Signal: "Speed",
				},
			},
		}

		err := config.Validate()
		a.NotNil(err)
		a.Contains(err.Error(), "duplicate signal topic")
	})

	t.Run("should return error for invalid message", func(t *testing.T) {
		a := assert.New(t)

		config := &Config{
			Messages: []Message{
				{
					ID:          123,
					Name:        "EngineSpeed",
					Length:      65, // Invalid: > 64 bits
					Transmitter: "Engine",
				},
			},
		}

		err := config.Validate()
		a.NotNil(err)
		a.Contains(err.Error(), "message Nº0")
	})
}

func TestMessageValidate(t *testing.T) {
	t.Run("should validate message with valid length", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{
			ID:          123,
			Name:        "EngineSpeed",
			Length:      8,
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

	t.Run("should return error when message length > 64", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{
			ID:          123,
			Name:        "EngineSpeed",
			Length:      65,
			Transmitter: "Engine",
		}

		err := message.Validate()
		a.Equal(ErrorMessageLengthOutOfBounds, err)
	})

	t.Run("should return error when signal lengths exceed message length", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{
			ID:          123,
			Name:        "EngineSpeed",
			Length:      4,
			Transmitter: "Engine",
			Signals: []Signal{
				{
					Name:     "Speed",
					StartBit: 0,
					Length:   3,
					Factor:   0.1,
				},
				{
					Name:     "Temp",
					StartBit: 3,
					Length:   3,
					Factor:   1.0,
				},
			},
		}

		err := message.Validate()
		a.Equal(ErrorSignalLengthsGreaterThanMessageLegnth, err)
	})

	t.Run("should return error for invalid signal", func(t *testing.T) {
		a := assert.New(t)

		message := &Message{
			ID:          123,
			Name:        "EngineSpeed",
			Length:      8,
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

		err := message.Validate()
		a.NotNil(err)
		a.Contains(err.Error(), "signal Nº0")
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
		a.Equal(ErrorSignalStartBitOutOfBounds, err)
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
		a.Equal(ErrorSignalLengthOutOfBounds, err)
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
		a.Equal(ErrorSignalFactorIsZero, err)
	})

	t.Run("should validate signal with correct integer and decimal figures", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{
			Name:           "Speed",
			StartBit:       0,
			Length:         16,
			IntegerFigures: 8,
			DecimalFigures: 8,
			Factor:         0.1,
		}

		err := signal.Validate()
		a.Nil(err)
	})

	t.Run("should return error when integer+decimal figures != length", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{
			Name:           "Speed",
			StartBit:       0,
			Length:         16,
			IntegerFigures: 4,
			DecimalFigures: 4,
			Factor:         0.1,
		}

		err := signal.Validate()
		a.Equal(ErrorSignalFigures, err)
	})

	t.Run("should allow zero figures when both are zero", func(t *testing.T) {
		a := assert.New(t)

		signal := &Signal{
			Name:           "Speed",
			StartBit:       0,
			Length:         2,
			IntegerFigures: 0,
			DecimalFigures: 0,
			Factor:         0.1,
		}

		err := signal.Validate()
		a.Nil(err)
	})
}

func TestSignalTopicValidate(t *testing.T) {
	t.Run("should validate topic with valid signal and topic", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			Topic:  "vehicle/engine/speed",
			Signal: "Speed",
		}

		err := topic.Validate()
		a.Nil(err)
	})

	t.Run("should return error when signal is empty", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			Topic:  "vehicle/engine/speed",
			Signal: "",
		}

		err := topic.Validate()
		a.Equal(ErrorInvalidSignalTopic, err)
	})

	t.Run("should return error when topic is empty", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			Topic:  "",
			Signal: "Speed",
		}

		err := topic.Validate()
		a.Equal(ErrorInvalidSignalTopic, err)
	})

	t.Run("should return error when both are empty", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			Topic:  "",
			Signal: "",
		}

		err := topic.Validate()
		a.Equal(ErrorInvalidSignalTopic, err)
	})
}
