package config

import (
	"encoding/json"
	"fmt"
)

const (
	// ParameterRequired indicates that the parameter is required, and must be
	// provided by the user.
	ParameterRequired parameterKind = iota + 1

	// ParameterOptional indicates that the parameter is optional, and may be
	// provided by the user.
	ParameterOptional

	// ParameterHardcoded indicates that the parameter is a hardcoded constant,
	// and must not be provided by the user.
	ParameterHardcoded
)

var _ json.Unmarshaler = (*parameterKind)(nil)

type parameterKind int

// UnmarshalJSON implements json.Unmarshaler.
func (p *parameterKind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case "optional":
		*p = ParameterOptional
	case "hardcoded":
		*p = ParameterHardcoded
	case "required", "":
		*p = ParameterRequired
	default:
		return fmt.Errorf("unknown parameter type %q", s)
	}

	return nil
}
