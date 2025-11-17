// Package utils contains methods to validate user input for cluster operations.
package utils

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

// ValidateClusterName accepts a cluster name and returns an error if it does not comply with the requirements.
func ValidateClusterName(name string) error {
	if len(name) < 3 {
		return errors.New("cluster name too short")
	}
	if len(name) > 28 {
		return errors.New("cluster name too long")
	}

	match, err := regexp.MatchString(`^(?:[a-z](?:[-a-z0-9]{1,26}[a-z0-9]))$`, name)
	if err != nil {
		return err
	}
	if !match {
		return errors.New(
			"the name does not match the requirements. " +
				"Only lowercase letters, numbers, and '-' allowed, must start with a letter and end with a letter or number. " +
				"A minimum length of 3 characters and a maximum length of 28 is allowed")
	}

	return nil
}

// ValidateLifespan accepts a duration and returns an error if it does not comply with the requirements.
func ValidateLifespan(lifespan time.Duration) error {
	if err := durationpb.New(lifespan).CheckValid(); err != nil {
		return fmt.Errorf("bad lifespan argument %q", err)
	}

	return checkValidLifespan(lifespan)
}

// This stub can contain custom logic to further validate lifespans.
func checkValidLifespan(_ time.Duration) error {
	return nil
}

// ValidateParameterArgument checks that key-value parameter arguments comply with the requirements.
func ValidateParameterArgument(parts []string) error {
	if len(parts) != 2 {
		return errors.New("must be of form key=value")
	}
	key, value := parts[0], parts[1]

	if key == "" {
		return errors.New("key is empty")
	}
	match, err := regexp.MatchString(`^[a-z0-9-]+$`, key)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("key is invalid format")
	}

	if value == "" {
		return errors.New("value is empty")
	}
	match, err = regexp.MatchString(`^[a-zA-Z0-9:\/\?._@-]+$`, value)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("value is invalid format")
	}

	return nil
}
