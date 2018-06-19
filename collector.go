package collector

import (
	"bytes"
	"context"
	"errors"
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

type Config struct {
	server string
	depth int
	packages []string
}

func (c *Config) checkOptions() error {
	if c.server == "" {
		return errors.New("OptionAgentURL is required")
	}

	if len(c.packages) == 0 {
		return errors.New("At least 1 package must be watched (OptionWatchPackage)")
	}

	return nil
}

type Option func(*Config)

func OptionWatchPackage(pkg string) Option {
	return func(c *Config) {
		c.packages = append(c.packages, pkg)
	}
}

func OptionAgentURL(url string) Option {
	return func(c *Config) {
		c.server = url
	}
}

func OptionCacheDepth(depth int) Option {
	return func(c *Config) {
		c.depth = depth
	}
}

func StartMindsightCollector(ctx context.Context, options ...Option) error {
	config := new(Config)
	for _, opt := range options {
		opt(config)
	}

	if err := config.checkOptions(); err != nil {
		return err
	}

	watched := make(map[string]interface{})

	for _, p := range config.packages {
		watched[p] = nil
	}

	watchedRadixTree := radix.NewFromMap(watched)
	cache := newSampleCache(config.depth, config.server)

	go sampleLoop(ctx, watchedRadixTree, cache)
	return nil
}
