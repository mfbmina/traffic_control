package circuitbreaker

import (
	"net/http"
)

const OPEN_STATE = "OPEN"
const CLOSED_STATE = "CLOSED"
const HALF_OPEN_STATE = "HALF_OPEN"
const DEFAULT_THRESHOLD = 5

type CircuitBreaker struct {
	Threshold    int
	FailureCount int
	State        string
	FailureFunc  func(r *http.Response) bool
}

func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		Threshold:    DEFAULT_THRESHOLD,
		FailureFunc:  defaultFailureFunc,
		FailureCount: 0,
		State:        CLOSED_STATE,
	}
}

func (cb *CircuitBreaker) Run(r *http.Request) {
	resp, err := http.DefaultClient.Do(r)

	if err != nil || cb.FailureFunc(resp) {
		cb.markFailure()
	}

	return
}

func (cb *CircuitBreaker) Reset() {
	cb.FailureCount = 0
	cb.State = CLOSED_STATE
}

func (cb *CircuitBreaker) markFailure() {
	cb.FailureCount += 1

	if cb.FailureCount >= cb.Threshold {
		cb.State = OPEN_STATE
	}
}

func defaultFailureFunc(r *http.Response) bool {
	// Default failure function
	defer r.Body.Close()

	if r.StatusCode >= 400 {
		return true
	}

	return false
}
