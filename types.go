package vera

type Config struct {
	Messages []Message
	Topics   []SignalTopic
}

type Node string

type Endianness uint

const (
	LittleEndian Endianness = iota
	BigEndian
)

type SignalTopic struct {
	Topic  string
	Signal string
}
