package collector

import (
	"bytes"
	"context"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/armon/go-radix"
	"github.com/maruel/panicparse/stack"
)

type nullWriter struct{}

func (w nullWriter) Write(buf []byte) (int, error) {
	return len(buf), nil
}

func sampleLoop(ctx context.Context, watched *radix.Tree, cache *sampleCache) {
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
					if err := cache.recordSample(c.Func.Raw); err != nil {
						log.Println("Error recording Mindsight sample:", err)
					}
					break
				}
			}
		}
	}
}

func StartMindsightCollector(ctx context.Context, server string, packages []string) {
	depth := DEFAULT_CACHE_DEPTH
	depthEnv := os.Getenv("MINDSIGHT_SAMPLE_CACHE_SIZE")

	if depthEnv != "" {
		// don't care if it doesn't parse, 0 it out
		depth, _ = strconv.Atoi(depthEnv)
	}

	watched := make(map[string]interface{})

	for _, p := range packages {
		watched[p] = nil
	}

	watchedRadixTree := radix.NewFromMap(watched)
	cache := newSampleCache(depth, server)

	go sampleLoop(ctx, watchedRadixTree, cache)
}
