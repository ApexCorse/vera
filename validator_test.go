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
					DLC:         9, // Invalid: > 8 bytes
					Transmitter: "Engine",
				},
			},
		}

		err := config.Validate()
		a.NotNil(err)
		a.Contains(err.Error(), "message Nº0")
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
		a.Error(err)
	})

	t.Run("should return error when topic is empty", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			Topic:  "",
			Signal: "Speed",
		}

		err := topic.Validate()
		a.Error(err)
	})

	t.Run("should return error when both are empty", func(t *testing.T) {
		a := assert.New(t)

		topic := &SignalTopic{
			Topic:  "",
			Signal: "",
		}

		err := topic.Validate()
		a.Error(err)
	})
}
