package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/armon/go-radix"
	"github.com/maruel/panicparse/stack"
)

type nullWriter struct{}

func (w nullWriter) Write(buf []byte) (int, error) {
	return len(buf), nil
}

func sampleLoop(ctx context.Context, watched *radix.Tree) {
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

		for _, g := range stackCtx.Goroutines {
			for _, c := range g.Signature.Stack.Calls {
				if _, _, present := watched.LongestPrefix(c.Func.Raw); present {
					fmt.Println(c.Func.Raw)
					break
				}
			}
		}
	}
}

func StartMindsightCollector(ctx context.Context, packages []string) {
	watched := make(map[string]interface{})

	for _, p := range packages {
		watched[p] = nil
	}

	watchedRadixTree := radix.NewFromMap(watched)

	go sampleLoop(ctx, watchedRadixTree)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	StartMindsightCollector(ctx, []string{"main"})

	fmt.Scanln()
}
