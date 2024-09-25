package first

import (
	"context"
	"errors"
	"sync"

	multierror "github.com/justindfuller/go-multierror"
)

// ErrNothingToWaitOn occurs when you call First.Wait() before First.Do().
// If this is intentional, the error is safe to ignore.
// The error is provided as a sentinel error so you can check for it.
//
// Example:
//
//	_, err := f.Wait()
//	if errors.Is(err, first.ErrNothingToWaitOn) {
//		// safe to continue?
//	}
var ErrNothingToWaitOn = errors.New("First.Wait() called without anything to wait on")

// First returns the first non-error result.
// Think of it like a sync.WaitGroup, except it stops waiting after the first group completes.
// You can also think of it like an errgroup.Group, except in the error scenario, it waits for all errors before completing.
//
// First uses generics to provide type-safe responses.
//
// Example:
//
//	var f first.First[*example]
//	var f first.First[mySampleStruct]
//	var f first.First[int64]
//
// You may use any type you need and it will be available to return from first.Do() and first.Wait().
//
// The zero value of First is ready to use, without further initialization.
// First should not be copied after first use.
// First is safe to use concurrently across multiple goroutines.
type First[T any] struct {
	mut     sync.Mutex
	errors  chan error
	result  chan T
	count   int
	context context.Context
	cancel  context.CancelFunc
}

// Do executes the provided function in a goroutine.
// It works in tandem with Wait() to retrieve the first result.
//
// When returning, the error should only have a value if T does not.
// If the error is non-nil, T is ignored.
// Do does not inspect the value of T. So, if error is nil, T is returned.
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

// DoContext works like Do, except it accepts and provides a context.
// The FIRST context provided to DoContext will be used. The rest will be ignored.
// After the first Do or DoContext call completes, the ctx provided to all DoContext callbacks will be canceled.
// This is useful for canceling long-running tasks that should short-circuit when the first operation completes.
// You are allowed to mix DoContext and Do with a single call to Wait.
//
// Example:
//
//	var f first.First
//
//	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
//		// do some long-running task that requires context
//		data, err := getFromDatabase(ctx)
//		if err != nil {
//			return nil, err
//		}
//		return data, nil
//	})
//
//	data, err := f.Wait()
func (f *First[T]) DoContext(ctx context.Context, fn func(context.Context) (T, error)) {
	f.mut.Lock()
	defer f.mut.Unlock()

	f.count++

	if f.result == nil {
		f.result = make(chan T, 1)
	}

	if f.errors == nil {
		f.errors = make(chan error)
	}

	if f.context == nil {
		// Avoid a panic with nil context
		if ctx == nil {
			ctx = context.Background()
		}

		f.context, f.cancel = context.WithCancel(ctx)
	}

	go func() {
		res, err := fn(f.context)
		if err != nil {
			f.errors <- err

			return
		}

		f.result <- res
	}()
}

// Wait for the first result or all errors.
//
// If you call Wait before Do, you will receive the ErrNothingToWaitOn error.
//
// Wait will block until a call to Do returns a nil error OR until all functions return a non-nil error.
// Neither Do nor Wait inspects the value of T, so any nil error value will result in Wait returning the value of T.
//
// Example:
//
//	res, err := t.Wait()
//	if err != nil {
//		// all calls to Do() returned an error.
//	}
//
//	fmt.Println(res) // the first value returned by any call to Do().
//
// You cannot call Do or DoContext once you call Wait. If you do, their calls will be blocked until after Wait completes.
func (f *First[T]) Wait() (T, error) {
	f.mut.Lock()
	defer f.mut.Unlock()

	if f.cancel != nil {
		defer f.cancel()
	}

	if f.count == 0 {
		return *new(T), ErrNothingToWaitOn
	}

	var errors []error

	for {
		if l := len(errors); l > 0 && l == f.count {
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
