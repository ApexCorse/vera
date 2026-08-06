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
				},
			},
			Topics: []SignalTopic{
				{
					MessageID: 123,
					Topic:     "vehicle/engine/speed",
					Signal:    "Speed",
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
					MessageID: 123,
					Topic:     "",
					Signal:    "Speed",
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
					MessageID: 123,
					Topic:     "vehicle/engine/speed",
					Signal:    "Speed",
				},
				{
					MessageID: 123,
					Topic:     "vehicle/engine/rpm",
					Signal:    "Speed",
				},
			},
		}

		err := config.Validate()
		a.NotNil(err)
		a.Contains(err.Error(), "duplicate signal topic")
	})

	t.Run("should scope topics by message ID and signal name", func(t *testing.T) {
		a := assert.New(t)

		config := &Config{
			Messages: []Message{
				{ID: 123, DLC: 1, Signals: []Signal{{Name: "Speed", Length: 1, Factor: 1}}},
				{ID: 456, DLC: 1, Signals: []Signal{{Name: "Speed", Length: 1, Factor: 1}}},
			},
			Topics: []SignalTopic{
				{MessageID: 123, Signal: "Speed", Topic: "vehicle/engine/speed"},
				{MessageID: 456, Signal: "Speed", Topic: "vehicle/wheel/speed"},
			},
		}

		err := config.Validate()
		a.Nil(err)
		a.Equal("vehicle/engine/speed", config.Messages[0].Signals[0].Topic)
		a.Equal("vehicle/wheel/speed", config.Messages[1].Signals[0].Topic)
	})

	t.Run("should return error for invalid message", func(t *testing.T) {
		a := assert.New(t)

		config := &Config{
			Messages: []Message{
				{
					ID:          123,
					Name:        "EngineSpeed",
					DLC:         9, // Invalid: > 8 bytes
					Transmitter: "Engine",
				},
			},
		}

		err := config.Validate()
		a.NotNil(err)
		a.Contains(err.Error(), "message DLC must be a number between 1 and 8")
	})
}

func TestSignalTopicValidate(t *testing.T) {
	t.Run("should validate topic with valid signal and topic", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			MessageID: 123,
			Topic:     "vehicle/engine/speed",
			Signal:    "Speed",
		}

		err := topic.Validate()
		a.Nil(err)
	})

	t.Run("should return error when signal is empty", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			MessageID: 123,
			Topic:     "vehicle/engine/speed",
			Signal:    "",
		}

		err := topic.Validate()
		a.Error(err)
	})

	t.Run("should return error when topic is empty", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			MessageID: 123,
			Topic:     "",
			Signal:    "Speed",
		}

		err := topic.Validate()
		a.Error(err)
	})

	t.Run("should return error when both are empty", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			MessageID: 123,
			Topic:     "",
			Signal:    "",
		}

		err := topic.Validate()
		a.Error(err)
	})
}
