package vera

import (
	"strconv"
	"strings"
)

type Message struct {
	Name        string
	ID          uint32
	DLC         uint8
	Transmitter Node
	Signals     []Signal

	signalsTotalLength  uint8
	lineNumber          int
	signalsBitPositions [64]int
}

func (m *Message) Validate() error {
	if m.DLC > 8 {
		return errorAtLine(m.lineNumber, "message DLC must be a number between 1 and 8")
	}

	//TODO(lentscode): check for start bits out of bounds
	if m.signalsTotalLength > m.DLC*8 {
		return errorAtLine(m.lineNumber, "sum of signal lengths must be less than or equal to (message DLC * 8)")
	}

	current := -1
	for _, p := range m.signalsBitPositions {
		if p == -1 {
			continue
		}

		if current == -1 {
			current = p
			continue
		}

		if current != p {
			return errorAtLine(m.lineNumber, "signals '%s' and '%s' cannot overlap", m.Signals[current].Name, m.Signals[p].Name)
		}

		current = -1
	}

	if current != -1 {
		return errorAtLine(m.lineNumber, "signal '%s' gives problems", m.Signals[current].Name)
	}

	for _, s := range m.Signals {
		if err := s.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func NewMessageFromLines(lines []string, startLineNumber int) (*Message, error) {
	message := &Message{
		lineNumber: startLineNumber,
	}

	for i := range message.signalsBitPositions {
		message.signalsBitPositions[i] = -1
	}

	if !strings.HasPrefix(lines[0], "BO_") {
		return nil, errorAtLine(message.lineNumber, "message line does not start with 'BO_'")
	}

	messageDefinition := lines[0]
	if err := message.parseDefinition(messageDefinition); err != nil {
		return nil, err
	}

	if len(lines) == 1 {
		return message, nil
	}

	for i, line := range lines[1:] {
		signal, err := NewSignalFromLine(message, line, startLineNumber+1+i)
		if err != nil {
			return nil, err
		}

		message.Signals = append(message.Signals, *signal)
	}

	return message, nil
}

func (m *Message) parseDefinition(line string) error {
	messageDefinitionParts := strings.Fields(line)
	if len(messageDefinitionParts) != 5 {
		return errorAtLine(m.lineNumber, `message line definition must be composed of 5 elements:
BO_ <MessageID> <MessageName>: <DLC> <TransmitterNode>`)
	}

	messageIDStr := messageDefinitionParts[1]
	if err := m.parseID(messageIDStr); err != nil {
		return err
	}

	messageName := messageDefinitionParts[2]
	if !strings.HasSuffix(messageName, ":") {
		return errorAtLine(m.lineNumber, "message name must end with a ':'")
	}
	messageName = strings.TrimSuffix(messageName, ":")

	dlcStr := messageDefinitionParts[3]
	dlc, err := strconv.Atoi(dlcStr)
	if err != nil {
		return errorAtLine(m.lineNumber, "message DLC must be a base 10 integer")
	}

	transmitter := messageDefinitionParts[4]

	m.Name = messageName
	m.DLC = uint8(dlc)
	m.Transmitter = Node(transmitter)

	return nil
}

func (m *Message) parseID(messageIDStr string) error {
	if !strings.HasPrefix(messageIDStr, "0x") {
		messageID, err := strconv.Atoi(messageIDStr)
		if err != nil {
			return errorAtLine(m.lineNumber, "message ID must be a base 10 or hexadecimal integer")
		}

		m.ID = uint32(messageID)

		return nil
	}

	messageIDHex := strings.TrimPrefix(messageIDStr, "0x")
	messageID, err := strconv.ParseUint(messageIDHex, 16, 32)
	if err != nil {
		return errorAtLine(m.lineNumber, "message ID must be a base 10 or hexadecimal integer")
	}

	m.ID = uint32(messageID)

	return nil
}
