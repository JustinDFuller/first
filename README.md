# first

[![Go Reference](https://pkg.go.dev/badge/github.com/justindfuller/first.svg)](https://pkg.go.dev/github.com/justindfuller/first)
[![Build Status](https://github.com/JustinDFuller/first/actions/workflows/build.yml/badge.svg)](https://github.com/JustinDFuller/first/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/justindfuller/first)](https://goreportcard.com/report/github.com/justindfuller/first)
![Go Test Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/justindfuller/63d4999a653a0555c9806062b40c0139/raw/first_coverage.json)

Get the first result without an error.

## Usage

First, install the package.

```
go get github.com/justindfuller/first
```

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
