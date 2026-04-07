package cli

import (
	"slices"
	"fmt"
)

func Contains(msg string, values  ...string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("nil value")
		}

		valid := slices.Contains(values, s)

		if !valid {
			return fmt.Errorf("%s", msg)
		}

		return nil
	}
}
