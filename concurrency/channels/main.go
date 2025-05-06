package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	ch := make(chan string)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			ch <- text
		}
	}()

	go func() {
		for text := range ch {
			time.Sleep(time.Second * 3)
			if n, err := strconv.Atoi(text); err == nil {
				fmt.Println(n * 2)
			} else {
				fmt.Println("not a number")
			}
		}
	}()

	select {}
}
