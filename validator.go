package vera

import "fmt"

func (c *Config) Validate() error {
	for i := range c.Messages {
		if err := c.Messages[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (m SignalMetadata) Validate() error {
	if m.StaleAfterMs != nil && *m.StaleAfterMs == 0 {
		return fmt.Errorf("stale-after milliseconds must be greater than zero")
	}
	if m.CriticalLow != nil && m.WarningLow != nil && *m.CriticalLow > *m.WarningLow {
		return fmt.Errorf("critical low must be less than or equal to warning low")
	}
	if m.WarningLow != nil && m.WarningHigh != nil && *m.WarningLow > *m.WarningHigh {
		return fmt.Errorf("warning low must be less than or equal to warning high")
	}
	if m.WarningHigh != nil && m.CriticalHigh != nil && *m.WarningHigh > *m.CriticalHigh {
		return fmt.Errorf("warning high must be less than or equal to critical high")
	}

	return nil
}
