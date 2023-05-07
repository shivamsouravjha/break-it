package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type CircuitBreaker struct {
	failureThreshold     int
	recoveryTime         time.Duration
	mutex                sync.Mutex
	consecutiveFail      int
	state                State
	lastStateTransition  time.Time
	lastSuccessfulInvoke time.Time
}

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func NewCircuitBreaker(failureThreshold int, recoveryTime time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold:     failureThreshold,
		recoveryTime:         recoveryTime,
		state:                Closed,
		lastStateTransition:  time.Now(),
		lastSuccessfulInvoke: time.Now(),
	}
}

func (cb *CircuitBreaker) Execute(function func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case Closed:
		if cb.consecutiveFail >= cb.failureThreshold {
			cb.state = Open
			cb.lastStateTransition = time.Now()
			return errors.New("circuit breaker is open")
		}

		err := function()
		if err != nil {
			cb.consecutiveFail++
			return err
		}

		cb.consecutiveFail = 0
		cb.lastSuccessfulInvoke = time.Now()
		return nil

	case Open:
		if time.Since(cb.lastStateTransition) >= cb.recoveryTime {
			cb.state = HalfOpen
			cb.consecutiveFail = 0
			return nil
		}

		return errors.New("circuit breaker is open")

	case HalfOpen:
		timeout := time.After(cb.recoveryTime)

		resultCh := make(chan error, 1)
		go func() {
			resultCh <- function()
		}()

		select {
		case result := <-resultCh:
			if result != nil {
				cb.lastStateTransition = time.Now()
				cb.state = Open
				cb.consecutiveFail++
				return errors.New("circuit breaker is open")
			}

			cb.lastStateTransition = time.Now()
			cb.state = Closed
			cb.consecutiveFail = 0
			cb.lastSuccessfulInvoke = time.Now()
			return nil

		case <-timeout:
			cb.lastStateTransition = time.Now()
			cb.state = Open
			cb.consecutiveFail++
			return errors.New("circuit breaker is open")
		}

	default:
		return errors.New("invalid circuit breaker state")
	}
}

func main() {
	circuitBreaker := NewCircuitBreaker(3, 5*time.Second)

	for i := 0; i < 10; i++ {
		err := circuitBreaker.Execute(func() error {
			fmt.Println("executing function...")
			return nil
		})

		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Function execution successful")
		}
	}
}
