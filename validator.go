package vera

import "fmt"

type signalTopicKey struct {
	messageID uint32
	signal    string
}

func (c *Config) Validate() error {
	topicsMap := make(map[signalTopicKey]string)
	for i, t := range c.Topics {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("topic Nº%d: %w", i, err)
		}

		key := signalTopicKey{messageID: t.MessageID, signal: t.Signal}
		if _, ok := topicsMap[key]; ok {
			return fmt.Errorf("duplicate signal topic for message %d, signal %s", t.MessageID, t.Signal)
		}
		topicsMap[key] = t.Topic
	}

	for i := range c.Messages {
		if err := c.Messages[i].Validate(); err != nil {
			return err
		}

		for j := range c.Messages[i].Signals {
			key := signalTopicKey{messageID: c.Messages[i].ID, signal: c.Messages[i].Signals[j].Name}
			if value, ok := topicsMap[key]; ok {
				c.Messages[i].Signals[j].Topic = value
			}
		}
	}

	return nil
}

func (t *SignalTopic) Validate() error {
	if t.Signal == "" || t.Topic == "" {
		return fmt.Errorf("signal topic must have a 'signal' name and a 'topic' name")
	}

	return nil
}
