package circuitbreaker

import "github.com/sony/gobreaker"

var (
	// MaxNumOfFailingRequests ...
	MaxNumOfFailingRequests = 10
	// FailingRatio ...
	FailingRatio = 0.6
)

// NewCircuitBreaker is a factory function returning a *gobreaker.CircuitBreaker
// with a default state-changing function that activates if the overall number
// of failing requests have reached a tweakable MaxNumOfFailingRequests cap and
// the failing ratio has met the FailingRatio.
func NewCircuitBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "circuitbreaker",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			ratio := float64(counts.TotalFailures) / float64(counts.Requests)
			return int(counts.Requests) > MaxNumOfFailingRequests && ratio >= FailingRatio
		},
	})
}
