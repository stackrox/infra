package utils

import (
	"log"
	"testing"
	"time"

	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

const defaultTimeout = 60 * time.Second

// AssertStatusBecomesWithin asserts that an infra cluster reaches a desired status within a defined time.
func AssertStatusBecomesWithin(t *testing.T, clusterID string, desiredStatus string, timeout time.Duration) {
	tick := 1 * time.Second
	conditionMet := func() bool {
		actualStatus, err := mock.InfractlGetStatusForID(clusterID)
		if err != nil {
			log.Printf("error when requesting status for cluster: '%s'\n", err.Error())
			return false
		}
		return desiredStatus == actualStatus
	}
	assert.Eventually(t, conditionMet, timeout, tick)
}

// AssertStatusBecomes asserts that an infra cluster eventually reaches a desired status.
func AssertStatusBecomes(t *testing.T, clusterID string, desiredStatus string) {
	AssertStatusBecomesWithin(t, clusterID, desiredStatus, defaultTimeout)
}

// AssertStatusRemainsFor asserts that an infra cluster remains in a desired status for a defined time.
func AssertStatusRemainsFor(t *testing.T, clusterID string, desiredStatus string, timeout time.Duration) {
	tick := 1 * time.Second
	conditionMet := func() bool {
		actualStatus, err := mock.InfractlGetStatusForID(clusterID)
		if err != nil {
			log.Printf("error when requesting status for cluster: '%s'\n", err.Error())
			return true
		}
		return desiredStatus != actualStatus
	}
	assert.Never(t, conditionMet, timeout, tick)
}
