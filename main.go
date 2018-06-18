package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/maruel/panicparse/stack"
)

type nullWriter struct{}

func (w nullWriter) Write(buf []byte) (int, error) {
	return len(buf), nil
}

func sampleLoop(ctx context.Context) {
	buf := make([]byte, 1<<16)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Millisecond):
			// take a stack sample
		}

		runtime.Stack(buf, true)
		stackCtx, err := stack.ParseDump(bytes.NewBuffer(buf), nullWriter{}, true)
		if err != nil {
			log.Println("parse stack error:", err)
		}

		fmt.Println(len(stackCtx.Goroutines))
	}
}

func StartMindsightCollector(ctx context.Context) {
	go sampleLoop(ctx)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	StartMindsightCollector(ctx)

	fmt.Scanln()
}
