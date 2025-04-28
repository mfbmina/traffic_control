package circuitbreaker

import (
	"errors"
	"time"
)

const OPEN_STATE = "OPEN"
const CLOSED_STATE = "CLOSED"
const HALF_OPEN_STATE = "HALF_OPEN"
const DEFAULT_THRESHOLD = 5
const DEFAULT_TIMEOUT = 5 * time.Second

var OpenErr = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	Failures         int
	FailureThreshold int
	OpenedAt         time.Time
	State            string
	Successes        int
	SuccessThreshold int
	Timeout          time.Duration
}

func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		FailureThreshold: DEFAULT_THRESHOLD,
		State:            CLOSED_STATE,
		SuccessThreshold: DEFAULT_THRESHOLD,
		Timeout:          DEFAULT_TIMEOUT,
	}
}

func (cb *CircuitBreaker) Close() {
	cb.State = CLOSED_STATE
}

func (cb *CircuitBreaker) HalfOpen() {
	cb.State = HALF_OPEN_STATE
	cb.Failures = 0
}

func (cb *CircuitBreaker) Open() {
	cb.State = OPEN_STATE
	cb.OpenedAt = time.Now()
}

func (cb *CircuitBreaker) Run(f func() (interface{}, error)) (interface{}, error) {
	if cb.State == OPEN_STATE && time.Since(cb.OpenedAt) < cb.Timeout {
		return nil, OpenErr
	}

	if cb.State == OPEN_STATE {
		cb.HalfOpen()
	}

	resp, err := f()
	if err != nil {
		cb.markFailure()
		return nil, err
	}

	cb.markSuccess()
	return resp, nil
}

func (cb *CircuitBreaker) WithFailureThreshold(t int) *CircuitBreaker {
	cb.FailureThreshold = t

	return cb
}

func (cb *CircuitBreaker) WithTimeout(t time.Duration) *CircuitBreaker {
	cb.Timeout = t

	return cb
}

func (cb *CircuitBreaker) WithSuccessThreshold(t int) *CircuitBreaker {
	cb.SuccessThreshold = t

	return cb
}

func (cb *CircuitBreaker) markFailure() {
	cb.Failures += 1
	if cb.Failures < cb.FailureThreshold {
		return
	}

	cb.Open()
}

func (cb *CircuitBreaker) markSuccess() {
	cb.Successes += 1
	if cb.State == CLOSED_STATE || cb.Successes < cb.SuccessThreshold {
		return
	}

	cb.Close()
}
