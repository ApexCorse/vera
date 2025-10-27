package vera

import "fmt"

func (c *Config) Validate() error {
	for i, m := range c.Messages {
		if err := m.Validate(); err != nil {
			return fmt.Errorf("message Nº%d: %s", i, err.Error())
		}
	}

	return nil
}

func (m *Message) Validate() error {
	if m.Length > 8 {
		return ErrorMessageLengthOutOfBounds
	}

	//TODO(lentscode): check for start bytes out of bounds
	totalLengths := uint8(0)

	for i, s := range m.Signals {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("signal Nº%d: %s", i, err.Error())
		}

		totalLengths += s.Length
	}
	if totalLengths > m.Length {
		return ErrorSignalLengthsGreaterThanMessageLegnth
	}

	return nil
}

func (s *Signal) Validate() error {
	if s.StartByte > 8 {
		return ErrorSignalStartByteOutOfBounds
	}
	if s.Length > 8 {
		return ErrorSignalLengthOutOfBounds
	}
	if s.Factor == 0 {
		return ErrorSignalFactorIsZero
	}

	if (s.IntegerFigures > 0 || s.DecimalFigures > 0) &&
		s.IntegerFigures+s.DecimalFigures != s.Length*8 {
		return ErrorSignalFigures
	}

	return nil
}
