package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var checks []func() error

func parse[T any](s string) (T, error) {
	var zero T

	switch any(zero).(type) {
	case int:
		v, err := strconv.Atoi(s)
		return any(v).(T), err

	case string:
		return any(s).(T), nil

	case bool:
		v, err := strconv.ParseBool(s)
		return any(v).(T), err

	case float64:
		v, err := strconv.ParseFloat(s, 64)
		return any(v).(T), err
	}

	return zero, fmt.Errorf("unsupported type")
}

func Flag[T any](name string, value T, usage string, validate func(T) error) (*T, func() *T) {
	var set bool

	flag.Func(name, usage, func(s string) error {
		v, err := parse[T](s)
		if err != nil {
			return err
		}

		if validate != nil {
			if err := validate(v); err != nil {
				return err
			}
		}

		value = v
		set = true
		return nil
	})

	required := func() *T {
		checks = append(checks, func() error {
			if !set {
				return fmt.Errorf("required argument: --%s", name)
			}
			return nil
		})
		return &value
	}

	return &value, required
}

func Required[T any](value *T, require func() *T) *T {
	return require()
}

func Validate() {
	for _, check := range checks {
		if err := check(); err != nil {
			fmt.Printf("%s\n", err.Error())
			flag.Usage()
			os.Exit(1)
		}
	}
}
