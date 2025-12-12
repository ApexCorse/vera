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
