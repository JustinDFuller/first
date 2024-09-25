package first_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/justindfuller/first"
)

type example struct {
	name string
}

var (
	errOne = errors.New("error one")
	errTwo = errors.New("error two")
)

func TestFirstSecond(t *testing.T) {
	var f first.First[*example]

	f.Do(func() (*example, error) {
		time.Sleep(10 * time.Millisecond)

		return &example{name: "one"}, nil
	})

	f.Do(func() (*example, error) {
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
}

func TestFirstFirst(t *testing.T) {
	var f first.First[*example]

	f.Do(func() (*example, error) {
		return &example{name: "one"}, nil
	})

	f.Do(func() (*example, error) {
		time.Sleep(10 * time.Millisecond)

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
}

func TestFirstError(t *testing.T) {
	var f first.First[*example]

	f.Do(func() (*example, error) {
		time.Sleep(10 * time.Millisecond)

		return &example{name: "one"}, nil
	})

	f.Do(func() (*example, error) {
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
}

func TestErrors(t *testing.T) {
	var f first.First[*example]

	f.Do(func() (*example, error) {
		time.Sleep(10 * time.Millisecond)

		return nil, errOne
	})

	f.Do(func() (*example, error) {
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
}

func TestFirstNone(t *testing.T) {
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
func TestFirstRoutines(t *testing.T) {
	var f first.First[*example]

	// Go ahead and set up one func to avoid the ErrNothingToWaitOn error.
	f.Do(func() (*example, error) {
		return &example{name: "one"}, nil
	})

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		f.Do(func() (*example, error) {
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
