package first

import (
	"errors"
	"sync"

	multierror "github.com/justindfuller/go-multierror"
)

var ErrNothingToWaitOn = errors.New("First.Wait() called without anything to wait on")

type First[T any] struct {
	mut    sync.Mutex
	errors chan error
	result chan T
	count  int
}

func (f *First[T]) Do(fn func() (T, error)) {
	f.mut.Lock()
	defer f.mut.Unlock()

	f.count++

	if f.result == nil {
		f.result = make(chan T, 1)
	}

	if f.errors == nil {
		f.errors = make(chan error)
	}

	go func() {
		res, err := fn()
		if err != nil {
			f.errors <- err

			return
		}

		f.result <- res
	}()
}

func (f *First[T]) Wait() (T, error) {
	f.mut.Lock()

	count := f.count

	f.mut.Unlock()

	if count == 0 {
		return *new(T), ErrNothingToWaitOn
	}

	var errors []error

	for {
		if l := len(errors); l > 0 && l == count {
			err := errors[0]

			for _, e := range errors[1:] {
				err = multierror.Join(err, e)
			}

			return *new(T), err
		}

		select {
		case res := <-f.result:
			return res, nil
		case err := <-f.errors:
			errors = append(errors, err)
		}
	}
}
