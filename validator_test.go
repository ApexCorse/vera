package vera

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func float32Pointer(value float32) *float32 { return &value }
func uint32Pointer(value uint32) *uint32    { return &value }

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		errText string
	}{
		{
			name: "valid metadata can omit every field",
			config: Config{Messages: []Message{{
				DLC: 1,
				Signals: []Signal{{
					Name: "Speed", Length: 1, Factor: 1,
				}},
			}}},
		},
		{
			name: "valid independent bounds",
			config: Config{Messages: []Message{{
				DLC: 1,
				Signals: []Signal{{
					Name: "Speed", Length: 1, Factor: 1,
					Metadata: SignalMetadata{
						CriticalLow:  float32Pointer(10),
						CriticalHigh: float32Pointer(5),
					},
				}},
			}}},
		},
		{
			name: "rejects zero stale after",
			config: Config{Messages: []Message{{
				DLC: 1,
				Signals: []Signal{{
					Name: "Speed", Length: 1, Factor: 1,
					Metadata: SignalMetadata{StaleAfterMs: uint32Pointer(0)},
				}},
			}}},
			errText: "stale-after milliseconds must be greater than zero",
		},
		{
			name: "rejects critical low above warning low",
			config: Config{Messages: []Message{{
				DLC: 1,
				Signals: []Signal{{
					Name: "Speed", Length: 1, Factor: 1,
					Metadata: SignalMetadata{
						CriticalLow: float32Pointer(11),
						WarningLow:  float32Pointer(10),
					},
				}},
			}}},
			errText: "critical low must be less than or equal to warning low",
		},
		{
			name: "rejects warning low above warning high",
			config: Config{Messages: []Message{{
				DLC: 1,
				Signals: []Signal{{
					Name: "Speed", Length: 1, Factor: 1,
					Metadata: SignalMetadata{
						WarningLow:  float32Pointer(11),
						WarningHigh: float32Pointer(10),
					},
				}},
			}}},
			errText: "warning low must be less than or equal to warning high",
		},
		{
			name: "rejects warning high above critical high",
			config: Config{Messages: []Message{{
				DLC: 1,
				Signals: []Signal{{
					Name: "Speed", Length: 1, Factor: 1,
					Metadata: SignalMetadata{
						WarningHigh:  float32Pointer(11),
						CriticalHigh: float32Pointer(10),
					},
				}},
			}}},
			errText: "warning high must be less than or equal to critical high",
		},
		{
			name:    "retains existing message validation",
			config:  Config{Messages: []Message{{DLC: 9}}},
			errText: "message DLC must be a number between 1 and 8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.errText == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errText)
		})
	}
}
