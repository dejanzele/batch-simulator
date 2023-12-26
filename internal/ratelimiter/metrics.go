package ratelimiter

import "fmt"

// Metrics returns the metrics of the rate limiter.
type Metrics struct {
	// Executed is the number of work items that have been processed.
	Executed int32
	// Failed is the number of work items that have failed.
	Failed int32
	// Succeeded is the number of work items that have succeeded.
	Succeeded int32
}

// Add adds the given metrics to the current metrics.
func (m *Metrics) Add(executed, failed, succeeded int32) {
	m.Executed += executed
	m.Failed += failed
	m.Succeeded += succeeded
}

func (m *Metrics) String() string {
	return fmt.Sprintf("(executed: %d, failed: %d, succeeded: %d)", m.Executed, m.Failed, m.Succeeded)
}
