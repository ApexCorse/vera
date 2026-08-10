package vera

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const legacyMQTTTopicCommentPrefix = "vera:mqtt-topic="

const (
	attributeMQTTTopic    = "VeraMqttTopic"
	attributeWarningLow   = "VeraWarningLow"
	attributeWarningHigh  = "VeraWarningHigh"
	attributeCriticalLow  = "VeraCriticalLow"
	attributeCriticalHigh = "VeraCriticalHigh"
	attributeStaleAfterMs = "VeraStaleAfterMs"
)

var signalAttributeTypes = map[string]string{
	attributeMQTTTopic:    "STRING",
	attributeWarningLow:   "FLOAT",
	attributeWarningHigh:  "FLOAT",
	attributeCriticalLow:  "FLOAT",
	attributeCriticalHigh: "FLOAT",
	attributeStaleAfterMs: "INT",
}

var (
	signalAttributeDefinitionPattern = regexp.MustCompile(`^\s*BA_DEF_\s+SG_\s+("(?:[^"\\]|\\.)*")\s+([A-Za-z]+)(?:\s+.*?)?\s*;\s*$`)
	signalAttributeAssignmentPattern = regexp.MustCompile(`^\s*BA_\s+("(?:[^"\\]|\\.)*")\s+SG_\s+(0x[0-9A-Fa-f]+|[0-9]+)\s+(\S+)\s+(.+?)\s*;\s*$`)
	signalAttributeNamePattern       = regexp.MustCompile(`^\s*BA_\s+("(?:[^"\\]|\\.)*")`)
)

type signalAttributeAssignment struct {
	name       string
	messageID  uint32
	signalName string
	value      string
	lineNumber int
}

func Parse(r io.Reader) (*Config, error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	config := &Config{}

	content := replaceNewLineCharacters(string(bytes))
	lines := strings.Split(content, "\n")
	definitions := make(map[string]struct{})
	assignments := make([]signalAttributeAssignment, 0)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if strings.Contains(trimmedLine, legacyMQTTTopicCommentPrefix) {
			return nil, errorAtLine(i, "legacy MQTT topic comments are not supported; use the %s signal attribute", attributeMQTTTopic)
		}

		switch {
		case strings.HasPrefix(trimmedLine, "BO_"):
			j := i + 1
			for ; j < len(lines); j++ {
				if !strings.HasPrefix(strings.TrimSpace(lines[j]), "SG_") {
					break
				}
			}

			message, err := NewMessageFromLines(lines[i:j], i)
			if err != nil {
				return nil, err
			}
			config.Messages = append(config.Messages, *message)
			i = j - 1
		case strings.HasPrefix(trimmedLine, "BA_DEF_"):
			name, attributeType, recognized, err := parseSignalAttributeDefinition(trimmedLine)
			if err != nil {
				return nil, errorAtLine(i, "%v", err)
			}
			if recognized {
				if _, ok := definitions[name]; ok {
					return nil, errorAtLine(i, "duplicate definition for signal attribute %s", name)
				}
				if attributeType != signalAttributeTypes[name] {
					return nil, errorAtLine(i, "signal attribute %s must use %s, got %s", name, signalAttributeTypes[name], attributeType)
				}
				definitions[name] = struct{}{}
			}
		case strings.HasPrefix(trimmedLine, "BA_"):
			assignment, recognized, err := parseSignalAttributeAssignment(trimmedLine, i)
			if err != nil {
				return nil, err
			}
			if recognized {
				if _, ok := definitions[assignment.name]; !ok {
					return nil, errorAtLine(i, "signal attribute %s requires a preceding BA_DEF_ SG_ declaration", assignment.name)
				}
				assignments = append(assignments, *assignment)
			}
		case strings.HasPrefix(trimmedLine, "BU_:"):
			nodesStr := strings.TrimPrefix(trimmedLine, "BU_:")
			for _, n := range strings.Fields(nodesStr) {
				config.Nodes = append(config.Nodes, Node(n))
			}
		case strings.HasPrefix(trimmedLine, "NS_ :"):
			j := i + 1
			for ; j < len(lines); j++ {
				line := strings.TrimSpace(lines[j])
				if line == "" {
					continue
				}
				if strings.Contains(line, " ") || strings.Contains(line, ":") {
					break
				}
				config.NewSymbols = append(config.NewSymbols, line)
			}
			i = j - 1
		}
	}

	if err := applySignalAttributeAssignments(config, assignments); err != nil {
		return nil, err
	}

	return config, nil
}

func parseSignalAttributeDefinition(line string) (name, attributeType string, recognized bool, err error) {
	matches := signalAttributeDefinitionPattern.FindStringSubmatch(line)
	if matches == nil {
		return "", "", false, nil
	}

	name, err = strconv.Unquote(matches[1])
	if err != nil {
		return "", "", false, fmt.Errorf("invalid signal attribute definition name: %w", err)
	}
	_, recognized = signalAttributeTypes[name]
	return name, matches[2], recognized, nil
}

func parseSignalAttributeAssignment(line string, lineNumber int) (*signalAttributeAssignment, bool, error) {
	matches := signalAttributeAssignmentPattern.FindStringSubmatch(line)
	if matches == nil {
		nameMatches := signalAttributeNamePattern.FindStringSubmatch(line)
		if nameMatches == nil {
			return nil, false, nil
		}
		name, err := strconv.Unquote(nameMatches[1])
		if err != nil {
			return nil, false, errorAtLine(lineNumber, "invalid signal attribute name: %v", err)
		}
		if _, ok := signalAttributeTypes[name]; ok {
			return nil, true, errorAtLine(lineNumber, "signal attribute %s has invalid assignment syntax", name)
		}
		return nil, false, nil
	}

	name, err := strconv.Unquote(matches[1])
	if err != nil {
		return nil, false, errorAtLine(lineNumber, "invalid signal attribute name: %v", err)
	}
	if _, ok := signalAttributeTypes[name]; !ok {
		return nil, false, nil
	}

	messageID, err := strconv.ParseUint(matches[2], 0, 32)
	if err != nil {
		return nil, true, errorAtLine(lineNumber, "invalid message ID in signal attribute %s: %v", name, err)
	}

	return &signalAttributeAssignment{
		name:       name,
		messageID:  uint32(messageID),
		signalName: matches[3],
		value:      strings.TrimSpace(matches[4]),
		lineNumber: lineNumber,
	}, true, nil
}

func applySignalAttributeAssignments(config *Config, assignments []signalAttributeAssignment) error {
	messages := make(map[uint32]*Message, len(config.Messages))
	for i := range config.Messages {
		messages[config.Messages[i].ID] = &config.Messages[i]
	}

	seen := make(map[string]struct{}, len(assignments))
	for _, assignment := range assignments {
		message, ok := messages[assignment.messageID]
		if !ok {
			return errorAtLine(assignment.lineNumber, "signal attribute %s references unknown message %d", assignment.name, assignment.messageID)
		}

		var signal *Signal
		for i := range message.Signals {
			if message.Signals[i].Name == assignment.signalName {
				signal = &message.Signals[i]
				break
			}
		}
		if signal == nil {
			return errorAtLine(assignment.lineNumber, "signal attribute %s references unknown signal %s on message %d", assignment.name, assignment.signalName, assignment.messageID)
		}

		key := fmt.Sprintf("%d\x00%s\x00%s", assignment.messageID, assignment.signalName, assignment.name)
		if _, ok := seen[key]; ok {
			return errorAtLine(assignment.lineNumber, "duplicate signal attribute %s for message %d, signal %s", assignment.name, assignment.messageID, assignment.signalName)
		}
		seen[key] = struct{}{}

		if err := setSignalMetadataValue(&signal.Metadata, assignment.name, assignment.value); err != nil {
			return errorAtLine(assignment.lineNumber, "invalid signal attribute %s: %v", assignment.name, err)
		}
	}
	for _, message := range config.Messages {
		for _, signal := range message.Signals {
			if err := signal.Metadata.Validate(); err != nil {
				return errorAtLine(signal.lineNumber, "signal metadata: %v", err)
			}
		}
	}

	return nil
}

func setSignalMetadataValue(metadata *SignalMetadata, name, rawValue string) error {
	switch name {
	case attributeMQTTTopic:
		value, err := parseDBCString(rawValue)
		if err != nil {
			return err
		}
		if value == "" {
			return fmt.Errorf("MQTT topic cannot be empty")
		}
		metadata.MQTTTopic = value
	case attributeWarningLow:
		value, err := parseDBCFloat(rawValue)
		if err != nil {
			return err
		}
		metadata.WarningLow = &value
	case attributeWarningHigh:
		value, err := parseDBCFloat(rawValue)
		if err != nil {
			return err
		}
		metadata.WarningHigh = &value
	case attributeCriticalLow:
		value, err := parseDBCFloat(rawValue)
		if err != nil {
			return err
		}
		metadata.CriticalLow = &value
	case attributeCriticalHigh:
		value, err := parseDBCFloat(rawValue)
		if err != nil {
			return err
		}
		metadata.CriticalHigh = &value
	case attributeStaleAfterMs:
		value, err := strconv.ParseUint(rawValue, 10, 32)
		if err != nil {
			return fmt.Errorf("must be an unsigned 32-bit integer: %w", err)
		}
		if value == 0 {
			return fmt.Errorf("must be greater than zero")
		}
		staleAfterMs := uint32(value)
		metadata.StaleAfterMs = &staleAfterMs
	}
	return nil
}

func parseDBCString(rawValue string) (string, error) {
	if len(rawValue) < 2 || rawValue[0] != '"' || rawValue[len(rawValue)-1] != '"' {
		return "", fmt.Errorf("must be a quoted string")
	}
	value, err := strconv.Unquote(rawValue)
	if err != nil {
		return "", fmt.Errorf("invalid quoted string: %w", err)
	}
	return value, nil
}

func parseDBCFloat(rawValue string) (float32, error) {
	value, err := strconv.ParseFloat(rawValue, 32)
	if err != nil {
		return 0, fmt.Errorf("must be a 32-bit floating-point value: %w", err)
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("must be finite")
	}
	return float32(value), nil
}
