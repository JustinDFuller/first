package first_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/justindfuller/first"
)

func ExampleFirst() {
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
		fmt.Println(err)
	}

	fmt.Println(res.name)
	// output: two
}

func ExampleFirst_Do() {
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
		fmt.Println(err)
	}

	fmt.Println(res.name)
	// output:
	// one
}

func ExampleFirst_Wait() {
	var f first.First[*example]

	f.Do(func() (*example, error) {
		time.Sleep(1 * time.Millisecond)
		return nil, errors.New("oops 2")
	})

	f.Do(func() (*example, error) {
		return nil, errors.New("oops")
	})

	res, err := f.Wait()
	if err != nil {
		fmt.Println(err)
	}

	if res != nil {
		fmt.Println(res)
	}
	// output:
	// Found 2 errors:
	//	oops
	//	oops 2
}

func ExampleFirst_DoContext() {
	ctx := context.Background()

	var wg sync.WaitGroup

	wg.Add(2)

	var f first.First[*example]

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()
		time.Sleep(1 * time.Millisecond)
		fmt.Printf("2 ctx=%s\n", ctx.Err())
		return nil, errors.New("oops 2")
	})

	f.DoContext(ctx, func(ctx context.Context) (*example, error) {
		defer wg.Done()
		fmt.Printf("1 ctx=%s\n", ctx.Err())
		return &example{name: "one"}, nil
	})

	_, _ = f.Wait()
	wg.Wait()
	// output:
	// 1 ctx=%!s(<nil>)
	// 2 ctx=context canceled
}
