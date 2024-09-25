package first_test

import (
	"errors"
	"fmt"
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
		fmt.Println("one")
		return &example{name: "one"}, nil
	})

	f.Do(func() (*example, error) {
		time.Sleep(10 * time.Millisecond)

		fmt.Println("two")
		return &example{name: "two"}, nil
	})

	res, err := f.Wait()
	fmt.Println("done")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res.name)
	// output:
	// one
	// done
	// one
}

func ExampleFirst_Wait() {
	var f first.First[*example]

	f.Do(func() (*example, error) {
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
