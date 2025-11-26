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

func (m *Message) Validate() error {
	if m.DLC > 8 {
		return ErrorMessageDLCOutOfBounds
	}

	//TODO(lentscode): check for start bits out of bounds
	totalLengths := uint8(0)

	for i, s := range m.Signals {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("signal Nº%d: %s", i, err.Error())
		}

		totalLengths += s.Length
	}
	if totalLengths > m.DLC*8 {
		return ErrorSignalLengthsGreaterThanMessageDLC
	}

	return nil
}

func (s *Signal) Validate() error {
	if s.StartBit >= 64 {
		return ErrorSignalStartBitOutOfBounds
	}
	if s.Length > 64 {
		return ErrorSignalLengthOutOfBounds
	}
	if s.Factor == 0 {
		return ErrorSignalFactorIsZero
	}

	if (s.IntegerFigures > 0 || s.DecimalFigures > 0) &&
		s.IntegerFigures+s.DecimalFigures != s.Length {
		return ErrorSignalFigures
	}

	return nil
}

func (t *SignalTopic) Validate() error {
	if t.Signal == "" || t.Topic == "" {
		return ErrorInvalidSignalTopic
	}

	return nil
}
