package vera

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const metadataDefinitions = `BA_DEF_ SG_ "VeraMqttTopic" STRING ;
BA_DEF_ SG_ "VeraWarningLow" FLOAT -1000000 1000000;
BA_DEF_ SG_ "VeraWarningHigh" FLOAT -1000000 1000000;
BA_DEF_ SG_ "VeraCriticalLow" FLOAT -1000000 1000000;
BA_DEF_ SG_ "VeraCriticalHigh" FLOAT -1000000 1000000;
BA_DEF_ SG_ "VeraStaleAfterMs" INT 1 4294967295;`

const oneSignalMessage = `BO_ 123 Engine: 8 Engine
 SG_ Speed : 0|8@1+ (1,0) [0|255] "km/h" Gateway`

func TestParseSignalMetadata(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		assert  func(*testing.T, *Config)
		errText string
	}{
		{
			name: "parses all signal metadata attributes",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraMqttTopic" SG_ 0x7b Speed "vehicle/engine/\"speed";
BA_ "VeraWarningLow" SG_ 123 Speed 10.5;
BA_ "VeraWarningHigh" SG_ 123 Speed 90;
BA_ "VeraCriticalLow" SG_ 123 Speed 5;
BA_ "VeraCriticalHigh" SG_ 123 Speed 100;
BA_ "VeraStaleAfterMs" SG_ 123 Speed 500;`,
			assert: func(t *testing.T, config *Config) {
				t.Helper()
				require.Len(t, config.Messages, 1)
				metadata := config.Messages[0].Signals[0].Metadata
				assert.Equal(t, `vehicle/engine/"speed`, metadata.MQTTTopic)
				require.NotNil(t, metadata.WarningLow)
				assert.Equal(t, float32(10.5), *metadata.WarningLow)
				require.NotNil(t, metadata.WarningHigh)
				assert.Equal(t, float32(90), *metadata.WarningHigh)
				require.NotNil(t, metadata.CriticalLow)
				assert.Equal(t, float32(5), *metadata.CriticalLow)
				require.NotNil(t, metadata.CriticalHigh)
				assert.Equal(t, float32(100), *metadata.CriticalHigh)
				require.NotNil(t, metadata.StaleAfterMs)
				assert.Equal(t, uint32(500), *metadata.StaleAfterMs)
			},
		},
		{
			name: "ignores unrelated attributes and defaults",
			input: `BA_DEF_ SG_ "Unrelated" STRING ;
BA_DEF_DEF_ "VeraMqttTopic" "default/topic";
` + oneSignalMessage + `
BA_ "Unrelated" SG_ 123 Speed "ignored";`,
			assert: func(t *testing.T, config *Config) {
				t.Helper()
				assert.Equal(t, SignalMetadata{}, config.Messages[0].Signals[0].Metadata)
			},
		},
		{
			name: "preserves ordinary comments and other DBC sections",
			input: `CM_ SG_ 123 Speed "Speed in km/h";
BU_: Gateway Engine
` + oneSignalMessage,
			assert: func(t *testing.T, config *Config) {
				t.Helper()
				assert.Len(t, config.Messages, 1)
				assert.Equal(t, []Node{"Gateway", "Engine"}, config.Nodes)
			},
		},
		{
			name: "rejects legacy topic comment",
			input: oneSignalMessage + `
CM_ SG_ 123 Speed "vera:mqtt-topic=vehicle/speed";`,
			errText: "legacy MQTT topic comments are not supported",
		},
		{
			name: "rejects malformed legacy topic comment",
			input: oneSignalMessage + `
CM_ SG_ 123 Speed vera:mqtt-topic=vehicle/speed`,
			errText: "legacy MQTT topic comments are not supported",
		},
		{
			name: "requires a preceding declaration",
			input: oneSignalMessage + `
BA_ "VeraMqttTopic" SG_ 123 Speed "vehicle/speed";`,
			errText: "requires a preceding BA_DEF_ SG_ declaration",
		},
		{
			name: "rejects wrong declaration type",
			input: `BA_DEF_ SG_ "VeraMqttTopic" INT 0 100;
` + oneSignalMessage,
			errText: "must use STRING, got INT",
		},
		{
			name: "rejects duplicate assignments",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraMqttTopic" SG_ 123 Speed "vehicle/one";
BA_ "VeraMqttTopic" SG_ 123 Speed "vehicle/two";`,
			errText: "duplicate signal attribute VeraMqttTopic",
		},
		{
			name: "rejects unknown message",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraMqttTopic" SG_ 999 Speed "vehicle/speed";`,
			errText: "references unknown message 999",
		},
		{
			name: "rejects unknown signal",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraMqttTopic" SG_ 123 Unknown "vehicle/speed";`,
			errText: "references unknown signal Unknown",
		},
		{
			name: "rejects empty MQTT topic",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraMqttTopic" SG_ 123 Speed "";`,
			errText: "MQTT topic cannot be empty",
		},
		{
			name: "rejects invalid known assignment syntax",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraMqttTopic" SG_ 123 Speed vehicle/speed;`,
			errText: "must be a quoted string",
		},
		{
			name: "rejects zero stale after",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraStaleAfterMs" SG_ 123 Speed 0;`,
			errText: "must be greater than zero",
		},
		{
			name: "rejects stale after overflow",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraStaleAfterMs" SG_ 123 Speed 4294967296;`,
			errText: "unsigned 32-bit integer",
		},
		{
			name: "rejects non-finite threshold",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraWarningLow" SG_ 123 Speed NaN;`,
			errText: "must be finite",
		},
		{
			name: "rejects incoherent threshold order",
			input: metadataDefinitions + "\n" + oneSignalMessage + `
BA_ "VeraCriticalLow" SG_ 123 Speed 20;
BA_ "VeraWarningLow" SG_ 123 Speed 10;`,
			errText: "critical low must be less than or equal to warning low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := Parse(strings.NewReader(tt.input))
			if tt.errText != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errText)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)
			tt.assert(t, config)
		})
	}
}

func TestParseBasicDBCSections(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		messageCount int
		nodes        []Node
		symbols      []string
	}{
		{
			name:         "empty input",
			input:        "",
			messageCount: 0,
		},
		{
			name: "messages nodes and symbols",
			input: `NS_ :
	BA_
	CM_
BU_: DriverGateway EngineGateway ABS
BO_ 123 EngineSpeed: 3 Engine
   SG_ EngineSpeed : 0|16@1+ (0.1,0) [0|8000] "RPM" DriverGateway
BO_ 0x124 Temperature: 1 Engine
   SG_ OilTemperature : 0|8@1- (1,-40) [-40|150] "degrees Celsius" DriverGateway`,
			messageCount: 2,
			nodes:        []Node{"DriverGateway", "EngineGateway", "ABS"},
			symbols:      []string{"BA_", "CM_"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := Parse(strings.NewReader(tt.input))
			require.NoError(t, err)
			assert.Len(t, config.Messages, tt.messageCount)
			assert.Equal(t, tt.nodes, config.Nodes)
			assert.Equal(t, tt.symbols, config.NewSymbols)
		})
	}
}
