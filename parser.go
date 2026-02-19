package vera

import (
	"fmt"
	"io"
	"strings"
)

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
		} else if strings.HasPrefix(lines[i], "TP_") {
			signalTopic, err := parseSignalTopic(lines[i])
			if err != nil {
				return nil, err
			}

			config.Topics = append(config.Topics, *signalTopic)
		}
	}

	return config, nil
}

func parseSignalTopic(topicLine string) (*SignalTopic, error) {
	lineParts := strings.Fields(topicLine)
	if len(lineParts) != 3 {
		return nil, fmt.Errorf(`signal topic has wrong structure: %s
Should be:
	TP_ <SignalName> <Topic>`, topicLine)
	}

	signalName := lineParts[1]
	topic := lineParts[2]
	topic = strings.TrimFunc(topic, func(r rune) bool {
		return r == '"'
	})

	signalTopic := &SignalTopic{
		Topic:  topic,
		Signal: signalName,
	}

	return signalTopic, nil
}
