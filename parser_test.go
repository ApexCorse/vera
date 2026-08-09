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

func TestParse_SignalWithSpacesInUnit(t *testing.T) {
	a := assert.New(t)

	config, err := Parse(strings.NewReader(`BO_ 0x123 Cabin: 8 Gateway
	SG_ InteriorTemperature: 0|8@1+ (0.5,-40) [-40|87.5] "degrees Celsius" Climate`))

	a.NoError(err)
	a.Len(config.Messages, 1)
	a.Len(config.Messages[0].Signals, 1)
	a.Equal("degrees Celsius", config.Messages[0].Signals[0].Unit)
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

func TestParse_WithNodesAndSymbols(t *testing.T) {
	t.Run("should parse config with nodes and symbols", func(t *testing.T) {
		a := assert.New(t)

		configStr := `NS_ :
	BA_
	CM_
	VAL_
BU_: DriverGateway EngineGateway ABS
BO_ 123 EngineSpeed: 3 Engine
   SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
	SG_ OilTemperature : 16|8@1- (1,-40) [-40|150] "ºC" DriverGateway,EngineGateway`

		reader := strings.NewReader(configStr)

		config, err := Parse(reader)
		a.Nil(err)
		a.NotNil(config)

		a.Len(config.Messages, 1)

		a.Len(config.Nodes, 3)
		a.Equal(Node("DriverGateway"), config.Nodes[0])
		a.Equal(Node("EngineGateway"), config.Nodes[1])
		a.Equal(Node("ABS"), config.Nodes[2])

		a.Len(config.NewSymbols, 3)
		a.Equal("BA_", config.NewSymbols[0])
		a.Equal("CM_", config.NewSymbols[1])
		a.Equal("VAL_", config.NewSymbols[2])
	})

	t.Run("should handle empty BU_ and NS_ sections", func(t *testing.T) {
		a := assert.New(t)

		configStr := `NS_ :
BU_: 
BO_ 123 EngineSpeed: 3 Engine
   SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway`

		reader := strings.NewReader(configStr)
		config, err := Parse(reader)

		a.Nil(err)
		a.NotNil(config)

		a.Len(config.Nodes, 0)
		a.Len(config.NewSymbols, 0)
		a.Len(config.Messages, 1)
	})

	t.Run("should skip unrelated DBC instructions and comments", func(t *testing.T) {
		a := assert.New(t)

		configStr := `CM_ "This is a comment"
NS_ :
	BA_
	CM_
BA_ "This is an attribute"
BU_: Gateway Engine
VAL_ 123 "This is a value"`

		reader := strings.NewReader(configStr)
		config, err := Parse(reader)

		a.Nil(err)
		a.NotNil(config)

		a.Len(config.Nodes, 2)
		a.Equal(Node("Gateway"), config.Nodes[0])
		a.Equal(Node("Engine"), config.Nodes[1])

		a.Len(config.NewSymbols, 2)
		a.Equal("BA_", config.NewSymbols[0])
		a.Equal("CM_", config.NewSymbols[1])
	})
}
