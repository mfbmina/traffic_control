package circuitbreaker

import (
	"net/http"
	"time"
)

const OPEN_STATE = "OPEN"
const CLOSED_STATE = "CLOSED"
const HALF_OPEN_STATE = "HALF_OPEN"
const DEFAULT_FAILURE_THRESHOLD = 5
const DEFAULT_TIMEOUT = 5 * time.Second

type CircuitBreaker struct {
	Failures         int
	FailureFunc      func(r *http.Response) bool
	FailureThreshold int
	OpenedAt         time.Time
	State            string
	Timeout          time.Duration
}

func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		FailureThreshold: DEFAULT_FAILURE_THRESHOLD,
		FailureFunc:      defaultFailureFunc,
		State:            CLOSED_STATE,
		Timeout:          DEFAULT_TIMEOUT,
	}
}

func (cb *CircuitBreaker) HalfOpen() {
	cb.State = HALF_OPEN_STATE
	cb.Failures = 0
}

func (cb *CircuitBreaker) Open() {
	cb.State = OPEN_STATE
	cb.OpenedAt = time.Now()
}

func (cb *CircuitBreaker) Run(r *http.Request) {
	if cb.State == OPEN_STATE && time.Since(cb.OpenedAt) < cb.Timeout {
		return
	}

	if cb.State == OPEN_STATE {
		cb.HalfOpen()
	}

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

func (cb *CircuitBreaker) WithTimeout(t time.Duration) *CircuitBreaker {
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
	if cb.Failures < cb.FailureThreshold {
		return
	}

	cb.Open()
}
