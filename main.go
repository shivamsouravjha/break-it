package main

import (
	"fmt"
	"sync"
	"time"
)

type CircuitBreaker struct {
	failureThreshold int
	recoveryTime     time.Duration
	mutex            sync.Mutex
	consecutiveFail  int
	state            State
}

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func NewCircuitBreaker(failureThreshold int, recoveryTime time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		recoveryTime:     recoveryTime,
		state:            Closed,
	}
}

func (cb *CircuitBreaker) Execute(function func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case Closed:
		if cb.consecutiveFail >= cb.failureThreshold {
			cb.state = Open
			go cb.autoReset()
			return fmt.Errorf("Circuit Breaker is OPEN")
		}

		err := function()
		if err != nil {
			cb.consecutiveFail++
		} else {
			cb.consecutiveFail = 0
		}

		return err

	case Open:
		return fmt.Errorf("Circuit Breaker is OPEN")

	case HalfOpen:
		err := function()
		if err != nil {
			cb.consecutiveFail++
			cb.state = Open
			go cb.autoReset()
		} else {
			cb.consecutiveFail = 0
			cb.state = Closed
		}

		return err

	default:
		return fmt.Errorf("Invalid Circuit Breaker state")
	}
}

func (cb *CircuitBreaker) autoReset() {
	time.Sleep(cb.recoveryTime)
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.state = HalfOpen
}

func main() {
	circuitBreaker := NewCircuitBreaker(3, 5*time.Second)

	for i := 0; i < 10; i++ {
		err := circuitBreaker.Execute(func() error {
			fmt.Println("Executing function...")
			if i%3 == 0 {
				return fmt.Errorf("Function execution failed")
			} else {
				return nil
			}
		})

		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Function execution successful")
		}
	}
}
