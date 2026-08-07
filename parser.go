package vera

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const mqttTopicCommentPrefix = "vera:mqtt-topic="

var signalCommentPattern = regexp.MustCompile(`^CM_\s+SG_\s+(0x[0-9A-Fa-f]+|[0-9]+)\s+(\S+)\s+"((?:[^"\\]|\\.)*)"\s*;\s*$`)

func Parse(r io.Reader) (*Config, error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	config := &Config{}

	content := string(bytes)
	content = replaceNewLineCharacters(content)

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return config, nil
	}

	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "BO_") {
			j := i + 1
			for ; j < len(lines); j++ {
				line := strings.TrimFunc(lines[j], func(r rune) bool {
					return r == ' ' || r == '\t'
				})
				if !strings.HasPrefix(line, "SG_") {
					break
				}
			}

			message, err := NewMessageFromLines(lines[i:j], i)
			if err != nil {
				return nil, err
			}

			config.Messages = append(config.Messages, *message)
			i = j - 1
		} else if strings.HasPrefix(lines[i], "CM_ SG_") {
			signalTopic, err := parseSignalTopicComment(lines[i])
			if err != nil {
				return nil, err
			}

			if signalTopic != nil {
				config.Topics = append(config.Topics, *signalTopic)
			}
		} else if strings.HasPrefix(lines[i], "BU_:") {
			nodesStr := strings.TrimPrefix(lines[i], "BU_:")
			nodeNames := strings.Fields(nodesStr)
			for _, n := range nodeNames {
				config.Nodes = append(config.Nodes, Node(n))
			}
		} else if strings.HasPrefix(lines[i], "NS_ :") {
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

	return config, nil
}

func parseSignalTopicComment(commentLine string) (*SignalTopic, error) {
	matches := signalCommentPattern.FindStringSubmatch(commentLine)
	if matches == nil {
		if strings.Contains(commentLine, mqttTopicCommentPrefix) {
			return nil, fmt.Errorf(`MQTT topic comment has wrong structure: %s
Should be:
	CM_ SG_ <MessageID> <SignalName> "vera:mqtt-topic=<Topic>";`, commentLine)
		}

		return nil, nil
	}

	comment, err := strconv.Unquote(`"` + matches[3] + `"`)
	if err != nil {
		return nil, fmt.Errorf("invalid signal comment: %w", err)
	}
	if !strings.HasPrefix(comment, mqttTopicCommentPrefix) {
		return nil, nil
	}

	messageID, err := strconv.ParseUint(matches[1], 0, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID in MQTT topic comment: %w", err)
	}

	signalTopic := &SignalTopic{
		MessageID: uint32(messageID),
		Topic:     strings.TrimPrefix(comment, mqttTopicCommentPrefix),
		Signal:    matches[2],
	}

	return signalTopic, nil
}
