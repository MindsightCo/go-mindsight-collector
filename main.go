package main

import (
	"fmt"
	"runtime"
	"time"
)

func printLoop() {
	time.Sleep(10 * time.Millisecond)

	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	fmt.Printf("%s", buf)

	printLoop()
}

func main() {

	go printLoop()

	fmt.Scanln()
}
