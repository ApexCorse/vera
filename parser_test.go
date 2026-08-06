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

func TestParseSignalTopicComment(t *testing.T) {
	t.Run("should parse a valid MQTT topic comment", func(t *testing.T) {
		a := assert.New(t)

		topic, err := parseSignalTopicComment(`CM_ SG_ 123 EngineSpeed "vera:mqtt-topic=vehicle/engine/speed";`)
		a.Nil(err)
		a.NotNil(topic)
		a.Equal(uint32(123), topic.MessageID)
		a.Equal("EngineSpeed", topic.Signal)
		a.Equal("vehicle/engine/speed", topic.Topic)
	})

	t.Run("should return an error for a malformed MQTT topic comment", func(t *testing.T) {
		a := assert.New(t)

		topic, err := parseSignalTopicComment(`CM_ SG_ 123 EngineSpeed "vera:mqtt-topic=vehicle/engine/speed"`)
		a.Error(err)
		a.Nil(topic)
		a.Contains(err.Error(), "MQTT topic comment has wrong structure")
	})

	t.Run("should ignore regular signal comments", func(t *testing.T) {
		a := assert.New(t)

		topic, err := parseSignalTopicComment(`CM_ SG_ 123 EngineSpeed "Engine speed in RPM";`)
		a.Nil(err)
		a.Nil(topic)
	})
}

func TestParse_WithTopics(t *testing.T) {
	t.Run("should parse config with topics", func(t *testing.T) {
		a := assert.New(t)

		configStr := `BO_ 123 EngineSpeed: 8 Engine
	SG_ Speed : 0|2@1+ (0.1,0) [0|100] "km/h" Gateway
CM_ SG_ 123 Speed "vera:mqtt-topic=vehicle/engine/speed";`
		reader := strings.NewReader(configStr)

		config, err := Parse(reader)
		a.Nil(err)
		a.NotNil(config)
		a.Len(config.Messages, 1)
		a.Len(config.Topics, 1)
		a.Equal(uint32(123), config.Topics[0].MessageID)
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

	t.Run("should skip non-topic DBC instructions and comments", func(t *testing.T) {
		a := assert.New(t)

		configStr := `CM_ "This is a comment"
BO_ 123 EngineSpeed: 8 Engine
	SG_ Speed : 0|2@1+ (0.1,0) [0|100] "km/h" Gateway
BA_ "AttributeName" "AttributeValue"
CM_ SG_ 123 Speed "vera:mqtt-topic=vehicle/engine/speed";`
		reader := strings.NewReader(configStr)

		config, err := Parse(reader)
		a.Nil(err)
		a.NotNil(config)
		a.Len(config.Messages, 1)
		a.Len(config.Topics, 1)
	})
}
