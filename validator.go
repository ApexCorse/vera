package vera

import "fmt"

func (c *Config) Validate() error {
	topicsMap := make(map[string]string)
	for i, t := range c.Topics {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("topic Nº%d: %w", i, err)
		}

		if _, ok := topicsMap[t.Signal]; ok {
			return fmt.Errorf("duplicate signal topic: %s", t.Topic)
		}
		topicsMap[t.Signal] = t.Topic
	}

	for i := range c.Messages {
		if err := c.Messages[i].Validate(); err != nil {
			return fmt.Errorf("message Nº%d: %s", i, err.Error())
		}

		for j := range c.Messages[i].Signals {
			if value, ok := topicsMap[c.Messages[i].Signals[j].Name]; ok {
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
