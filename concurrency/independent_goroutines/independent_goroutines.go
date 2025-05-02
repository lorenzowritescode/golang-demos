package main

import (
	"fmt"
	"math/rand"
)

// randomNumberPrinter prints its ID and a random number.
func randomNumberPrinter(id int) {
	randomNum := rand.Intn(100) 
	fmt.Printf("Goroutine %d: Random number is %d\n", id, randomNum)
}

func main() {
	fmt.Println("Main: Starting application.")

	numGoroutines := 10
	fmt.Printf("Main: Launching %d goroutines...\n", numGoroutines)

	for i := 1; i <= numGoroutines; i++ {
		// Launch each task as a goroutine (fire and forget)
		go randomNumberPrinter(i)
	}

	fmt.Println("Main: All goroutines launched.")

	fmt.Println("Main: Exiting application.")
}
