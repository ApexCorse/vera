package vera

type Config struct {
	Messages   []Message
	Nodes      []Node
	NewSymbols []string
}

type Node string

type Endianness uint

const (
	LittleEndian Endianness = iota
	BigEndian
)

// SignalMetadata is optional, signal-scoped configuration stored in DBC
// attributes. Its fields intentionally use pointers where zero is a valid
// configured value and must be distinguishable from an absent attribute.
type SignalMetadata struct {
	MQTTTopic    string
	WarningLow   *float32
	WarningHigh  *float32
	CriticalLow  *float32
	CriticalHigh *float32
	StaleAfterMs *uint32
}
