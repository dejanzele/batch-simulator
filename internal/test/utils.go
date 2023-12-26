package test

import (
	"os"
	"testing"
)

// IntegrationTest skips the test if INTEGRATION_TEST is not set to true.
func IntegrationTest(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("skipping integration test as INTEGRATION_TEST is not set to true")
	}
}
