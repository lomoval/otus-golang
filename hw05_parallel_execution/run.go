package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

var ErrIncorrectGoroutinesCount = errors.New("incorrect number of goroutines")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrIncorrectGoroutinesCount
	}
	if len(tasks) == 0 {
		return nil
	}

	done := make(chan struct{})
	tasksCh := make(chan Task)
	results := make(chan error)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasksCh {
				select {
				case results <- task():
				case <-done:
					return
				}
			}
		}()
	}

	go func() {
		for _, task := range tasks {
			select {
			case tasksCh <- task:
			case <-done:
				break
			}
		}
		close(tasksCh)
		wg.Wait()
		close(results)
	}()

	var errCount int
	for result := range results {
		if result != nil {
			errCount++
		}
		if errCount == m && m > 0 {
			close(done)
			wg.Wait()
			return ErrErrorsLimitExceeded
		}
	}
	return nil
}
