// Example file for demonstrating linting capabilities
package main

import (
	"fmt"
)

func main() {
	// Example of good code
	message := "Hello, World!"
	fmt.Println(message)

	// Check error handling
	result := doSomething()
	if result != nil {
		fmt.Printf("Error: %v\n", result)
	}
}

func doSomething() error {
	// Properly return nil for no error
	return nil
}
