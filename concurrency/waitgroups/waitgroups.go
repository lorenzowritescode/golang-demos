package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func randomNumberPrinter(id int, wg *sync.WaitGroup) {
	defer wg.Done() // Signal completion when the function returns
	randomNum := rand.Intn(100)
	fmt.Printf("Goroutine %d: Random number is %d\n", id, randomNum)
}

func main() {
	fmt.Println("Main: Starting application.")
	numGoroutines := 10

	var wg sync.WaitGroup // Create a WaitGroup
	wg.Add(numGoroutines) // Increment counter for the number of goroutines

	for i := 1; i <= numGoroutines; i++ {
		go randomNumberPrinter(i, &wg) // Launch goroutine, passing the WaitGroup pointer
	}

	wg.Wait() // Block until the WaitGroup counter is zero
	fmt.Println("Main: Exiting application.")
}
