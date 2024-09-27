# first

[![Go Reference](https://pkg.go.dev/badge/github.com/justindfuller/first.svg)](https://pkg.go.dev/github.com/justindfuller/first)
[![Build Status](https://github.com/JustinDFuller/first/actions/workflows/build.yml/badge.svg)](https://github.com/JustinDFuller/first/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/justindfuller/first)](https://goreportcard.com/report/github.com/justindfuller/first)
![Go Test Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/justindfuller/63d4999a653a0555c9806062b40c0139/raw/first_coverage.json)

## TL;DR

It is a concurrency tool that gets the first value _or_ all the errors.

## 📖 Table of Contents 📖
- [What?](#what)
- [When?](#when)
- [Usage](#usage)
- [Documentation](#documentation)

## What?

`First` is a synchronization tool. It gets the first result that does not return an error. Otherwise, it gets all the errors.

You might think of it as similar to [`sync.WaitGroup`](https://pkg.go.dev/sync#WaitGroup). Except, it either waits for the first result _or_ all of the errors.

You might think of it as similar to an [`errgroup.Group`](https://pkg.go.dev/golang.org/x/sync/errgroup#Group). Except, it does not wait for all functions to return. It waits for the first function to return a value _or_ it waits for all the functions to return an error.

## When?

One example is retrieving from a fast and slow data store concurrently. You might do this if you have no issue putting the full load on both resources.

Basically, you might use it any time you want to concurrently perform multiple tasks, but only need to wait for one of them to complete without error.

## Usage

### Install

First, install the package.

```
go get github.com/justindfuller/first
```

### Basic Example

Then, use it.

```go
package main

type example struct{
    name string
}

func main() {
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
		log.Fatalf("Error: %s", err)
	}
	
	log.Printf("Result: %v", res) // prints "two"
}
```

### Context Example

It also supports using contexts.

```go
package main

type example struct{
    name string
}

func main() {
	f, ctx := first.WithContext[*example](context.Background())
	
	f.Do(func() (*example, error) {
		select {
			case <-time.After(10 * time.Millisecond):		
				return &example{name: "one"}, nil
			case <-ctx.Done():
				log.Print("Skipped one")
				return nil, ctx.Err()
		}
	})
	
	f.Do(func() (*example, error) {
		select {
			case <-time.After(1 * time.Millisecond):		
				return &example{name: "two"}, nil
			case <-ctx.Done():
				log.Print("Skipped two")
				return nil, ctx.Err()
		}	
	})
	
	res, err := f.Wait()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	
	log.Printf("Result: %v", res) // prints "two"
	// Also prints, "skipped one"

	log.Printf("Context: %s", ctx.Err())
	// Prints: "Context: context canceled"
}
```

## Documentation

Please refer to the go documentation hosted on [pkg.go.dev](https://pkg.go.dev/github.com/justindfuller/first). You can see [all available types and methods](https://pkg.go.dev/github.com/justindfuller/first#pkg-index) and [runnable examples](https://pkg.go.dev/github.com/justindfuller/first#pkg-examples).
