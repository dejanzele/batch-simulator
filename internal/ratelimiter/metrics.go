package ratelimiter

import "fmt"

// Metrics returns the metrics of the rate limiter.
type Metrics struct {
	// Executed is the number of work items that have been processed.
	Executed int
	// Failed is the number of work items that have failed.
	Failed int
	// Succeeded is the number of work items that have succeeded.
	Succeeded int
}

// Add adds the given metrics to the current metrics.
func (m *Metrics) Add(executed, failed, succeeded int) {
	m.Executed += executed
	m.Failed += failed
	m.Succeeded += succeeded
}

func (m *Metrics) String() string {
	return fmt.Sprintf("(executed: %d, failed: %d, succeeded: %d)", m.Executed, m.Failed, m.Succeeded)
}
