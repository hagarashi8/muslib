package envbinder

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

var NoEnvVarError = fmt.Errorf("No value")
var InvalidBoolValue = fmt.Errorf("Invalid Value for Bool")

type EnvBinder struct {
	Err error
}

func (b *EnvBinder) String(v *string, name string) *EnvBinder {
	s, ok := os.LookupEnv(name)
	if !ok {
		b.Err = NoEnvVarError
		return b
	}
	*v = s
	return b
}

func (b *EnvBinder) StringOrDef(v *string, name string, def string) *EnvBinder {
	s, ok := os.LookupEnv(name)
	if !ok {
		*v = def
		return b
	}
	*v = s
	return b
}

func (b *EnvBinder) Bool(v *bool, name string) *EnvBinder {
	s, ok := os.LookupEnv(name)
	if !ok {
		b.Err = NoEnvVarError
		return b
	}
	s = strings.ToLower(s)
	if slices.Contains([]string{"yes", "enable", "enabled", "true"}, s) {
		*v = true
	} else if slices.Contains([]string{"no", "disable", "disabled", "false"}, s) {
		*v = false
	} else {
		b.Err = InvalidBoolValue
	}
	return b
}

func (b *EnvBinder) BoolOrDef(v *bool, name string, def bool) *EnvBinder {
	s, ok := os.LookupEnv(name)
	if !ok {
		*v = def
		return b
	}
	s = strings.ToLower(s)
	if slices.Contains([]string{"yes", "enable", "enabled", "true"}, s) {
		*v = true
	} else if slices.Contains([]string{"no", "disable", "disabled", "false"}, s) {
		*v = false
	} else {
		*v = def
	}
	return b
}

func (b *EnvBinder) Int(v *int, name string) *EnvBinder {
	s, ok := os.LookupEnv(name)
	if !ok {
		b.Err = NoEnvVarError
		return b
	}
	a, err := strconv.Atoi(s)
	if err != nil {
		b.Err = err
		return b
	}
	*v = a
	return b
}

func (b *EnvBinder) IntOrDef(v *int, name string, def int) *EnvBinder {
	s, ok := os.LookupEnv(name)
	if !ok {
		*v = def
		return b
	}
	a, err := strconv.Atoi(s)
	if err != nil {
		*v = def
		return b
	}
	*v = a
	return b
}

func (b *EnvBinder) BindError() error {
	return b.Err
}
