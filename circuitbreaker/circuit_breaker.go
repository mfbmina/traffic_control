package circuitbreaker

import (
	"net/http"
)

const OPEN_STATE = "OPEN"
const CLOSED_STATE = "CLOSED"
const HALF_OPEN_STATE = "HALF_OPEN"
const DEFAULT_FAILURE_THRESHOLD = 5
const DEFAULT_TIMEOUT = 5

type CircuitBreaker struct {
	Failures         int
	FailureFunc      func(r *http.Response) bool
	FailureThreshold int
	State            string
	Timeout          int
}

func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		FailureThreshold: DEFAULT_FAILURE_THRESHOLD,
		FailureFunc:      defaultFailureFunc,
		State:            CLOSED_STATE,
		Timeout:          DEFAULT_TIMEOUT,
	}
}

func (cb *CircuitBreaker) Reset() {
	cb.Failures = 0
	cb.State = CLOSED_STATE
}

func (cb *CircuitBreaker) Run(r *http.Request) {
	resp, err := http.DefaultClient.Do(r)

	if err != nil || cb.FailureFunc(resp) {
		cb.markFailure()
	}

	return
}

func (cb *CircuitBreaker) WithFailureFunc(f func(r *http.Response) bool) *CircuitBreaker {
	cb.FailureFunc = f

	return cb
}

func (cb *CircuitBreaker) WithFailureThreshold(t int) *CircuitBreaker {
	cb.FailureThreshold = t

	return cb
}

func (cb *CircuitBreaker) WithTimeout(t int) *CircuitBreaker {
	cb.Timeout = t

	return cb
}

func defaultFailureFunc(r *http.Response) bool {
	// Default failure function
	defer r.Body.Close()

	if r.StatusCode >= 400 {
		return true
	}

	return false
}

func (cb *CircuitBreaker) markFailure() {
	cb.Failures += 1

	if cb.Failures >= cb.FailureThreshold {
		cb.State = OPEN_STATE
	}
}
