package vera

type Config struct {
	Messages []Message
}

type Message struct {
	Name        string
	ID          uint32
	Length      uint8
	Transmitter Node
	Signals     []Signal
}

type Signal struct {
	Name      string
	StartByte uint8
	Length    uint8
	Endianness
	Signed         bool
	IntegerFigures uint8
	DecimalFigures uint8
	Factor         float32
	Offset         float32
	Min            float32
	Max            float32
	Unit           string
	Receivers      []Node
}

type Node string

type Endianness uint

const (
	LittleEndian Endianness = iota
	BigEndian
)
