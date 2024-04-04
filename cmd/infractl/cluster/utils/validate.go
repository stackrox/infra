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
			"The name does not match the requirements. " +
				"Only lowercase letters, numbers, and '-' allowed, must start with a letter and end with a letter or number. " +
				"A minimum length of 3 characters and a maximum length of 28 is allowed.")
	}

	return nil
}

// ValidateLifespan accepts a duration and returns an error if it does not comply with ther requirements.
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
