# first

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
