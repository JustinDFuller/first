package first_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/justindfuller/first"
)

func TestContextFirstSecond(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	ctx := context.Background()

	var f first.First[*example]

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		time.Sleep(10 * time.Millisecond)

		if ctx.Err() == nil {
			t.Fatal("Unexpected nil context")
		}

		return &example{name: "one"}, nil
	})

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		if ctx.Err() != nil {
			t.Fatalf("Unexpected context error: %s", ctx.Err())
		}

		return &example{name: "two"}, nil
	})

	res, err := f.Wait()
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if res == nil {
		t.Fatal("Expected non-nil res, got nil")
	}

	if res.name != "two" {
		t.Fatalf("Expected two, got %s", res.name)
	}

	wg.Wait()
}

func TestContextFirstFirst(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	ctx := context.Background()

	var f first.First[*example]

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		if ctx.Err() != nil {
			t.Fatalf("Unexpected context error: %s", ctx.Err())
		}

		return &example{name: "one"}, nil
	})

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		time.Sleep(10 * time.Millisecond)

		if ctx.Err() == nil {
			t.Fatal("Unexpected nil context")
		}

		return &example{name: "two"}, nil
	})

	res, err := f.Wait()
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if res == nil {
		t.Fatal("Expected non-nil res, got nil")
	}

	if res.name != "one" {
		t.Fatalf("Expected one, got %s", res.name)
	}

	wg.Wait()
}

func TestContextFirstError(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	ctx := context.Background()

	var f first.First[*example]

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		time.Sleep(10 * time.Millisecond)

		if ctx.Err() != nil {
			t.Fatalf("Unexpected context error: %s", ctx.Err())
		}

		return &example{name: "one"}, nil
	})

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		if ctx.Err() != nil {
			t.Fatalf("Unexpected context error: %s", ctx.Err())
		}

		return nil, errors.New("oops")
	})

	res, err := f.Wait()
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if res == nil {
		t.Fatal("Expected non-nil res, got nil")
	}

	if res.name != "one" {
		t.Fatalf("Expected two, got %s", res.name)
	}

	wg.Wait()
}

func TestContextErrors(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	ctx := context.Background()

	var f first.First[*example]

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		time.Sleep(10 * time.Millisecond)

		if ctx.Err() != nil {
			t.Fatalf("Unexpected context error: %s", ctx.Err())
		}

		return nil, errOne
	})

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		if ctx.Err() != nil {
			t.Fatalf("Unexpected context error: %s", ctx.Err())
		}

		return nil, errTwo
	})

	res, err := f.Wait()
	if err == nil {
		t.Fatal("Expected non-nil error, got nil")
	}

	if res != nil {
		t.Errorf("Expected nil res, got %v", res)
	}

	if !errors.Is(err, errOne) {
		t.Errorf("Expected %s, got %s", errOne, err)
	}

	if !errors.Is(err, errTwo) {
		t.Errorf("Expected %s, got %s", errTwo, err)
	}

	wg.Wait()
}

func TestContextFirstNone(t *testing.T) {
	var f first.First[*example]

	res, err := f.Wait()
	if res != nil {
		t.Errorf("Expected nil res, got %v", res)
	}

	if err == nil {
		t.Fatal("Expected non-nil error, got nil")
	}

	if !errors.Is(err, first.ErrNothingToWaitOn) {
		t.Fatalf("Expected %s, got %s", first.ErrNothingToWaitOn, err)
	}
}

// This test may seem a little odd.
// The goal is to ensure all First methods are safe to call within goroutines.
// If they are not, the -race flag in the go test command will report a race condition.
func TestContextFirstRoutines(t *testing.T) {
	ctx := context.Background()

	var f first.First[*example]

	// Go ahead and set up one func to avoid the ErrNothingToWaitOn error.
	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		if ctx.Err() != nil {
			t.Fatalf("Unexpected context error: %s", ctx.Err())
		}

		return &example{name: "one"}, nil
	})

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		f.DoContext(ctx, func(ctx context.Context) (*example, error) {
			time.Sleep(10 * time.Millisecond)

			return &example{name: "two"}, nil
		})
	}()

	var res *example
	var err error

	go func() {
		defer wg.Done()

		res, err = f.Wait()
	}()

	wg.Wait()

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if res == nil {
		t.Fatal("Expected non-nil res, got nil")
	}

	if res.name != "one" {
		t.Fatalf("Expected one, got %s", res.name)
	}
}

func TestRespectsCancel(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())

	var f first.First[*example]

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		wg.Wait()

		return &example{name: "one"}, nil
	})

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		wg.Wait()

		return &example{name: "two"}, nil
	})

	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	_, err := f.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected %s, got %s", context.Canceled, err)
	}
}

func TestRespectsWithContextCancel(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())

	f, ctx := first.WithContext[*example](ctx)

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		wg.Wait()

		return &example{name: "one"}, nil
	})

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()

		wg.Wait()

		return &example{name: "two"}, nil
	})

	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	_, err := f.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected %s, got %s", context.Canceled, err)
	}
}
