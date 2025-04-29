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

var ErrOpenCircuit = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	Failures         int
	FailureThreshold int
	OpenedAt         time.Time
	State            string
	CloseCheck       func() bool
	HalfOpenCheck    func() bool
	OpenCheck        func() bool
	Successes        int
	SuccessThreshold int
	Timeout          time.Duration
}

func New() *CircuitBreaker {
	cb := &CircuitBreaker{State: CLOSED_STATE}

	return cb.WithFailureThreshold(DEFAULT_THRESHOLD).
		WithSuccessThreshold(DEFAULT_THRESHOLD).
		WithTimeout(DEFAULT_TIMEOUT).
		WithCloseCheck(cb.defaultCloseCheck).
		WithHalfOpenCheck(cb.defaultHalfOpenCheck).
		WithOpenCheck(cb.defaultOpenCheck)
}

func (cb *CircuitBreaker) Close() {
	cb.State = CLOSED_STATE
}

func (cb *CircuitBreaker) HalfOpen() {
	cb.State = HALF_OPEN_STATE
	cb.Failures = 0
	cb.Successes = 0
}

func (cb *CircuitBreaker) Open() {
	cb.State = OPEN_STATE
	cb.OpenedAt = time.Now()
}

func (cb *CircuitBreaker) Run(f func() (interface{}, error)) (interface{}, error) {
	if cb.HalfOpenCheck() {
		cb.HalfOpen()
	}

	if cb.State == OPEN_STATE {
		return nil, ErrOpenCircuit
	}

	resp, err := f()
	if err != nil {
		cb.markFailure()
		return nil, err
	}

	cb.markSuccess()
	return resp, nil
}

func (cb *CircuitBreaker) WithCloseCheck(f func() bool) *CircuitBreaker {
	cb.CloseCheck = f

	return cb
}

func (cb *CircuitBreaker) WithFailureThreshold(t int) *CircuitBreaker {
	cb.FailureThreshold = t

	return cb
}

func (cb *CircuitBreaker) WithHalfOpenCheck(f func() bool) *CircuitBreaker {
	cb.HalfOpenCheck = f

	return cb
}

func (cb *CircuitBreaker) WithTimeout(t time.Duration) *CircuitBreaker {
	cb.Timeout = t

	return cb
}

func (cb *CircuitBreaker) WithOpenCheck(f func() bool) *CircuitBreaker {
	cb.OpenCheck = f

	return cb
}

func (cb *CircuitBreaker) WithSuccessThreshold(t int) *CircuitBreaker {
	cb.SuccessThreshold = t

	return cb
}

func (cb *CircuitBreaker) defaultCloseCheck() bool {
	return cb.State == HALF_OPEN_STATE && cb.Successes >= cb.SuccessThreshold
}

func (cb *CircuitBreaker) defaultHalfOpenCheck() bool {
	return cb.State == OPEN_STATE && time.Since(cb.OpenedAt) > cb.Timeout
}

func (cb *CircuitBreaker) defaultOpenCheck() bool {
	return cb.Failures >= cb.FailureThreshold
}

func (cb *CircuitBreaker) markFailure() {
	cb.Failures += 1

	if !cb.OpenCheck() {
		return
	}

	cb.Open()
}

func (cb *CircuitBreaker) markSuccess() {
	cb.Successes += 1
	if !cb.CloseCheck() {
		return
	}

	cb.Close()
}
