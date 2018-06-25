package collector

import (
	"bytes"
	"context"
	"errors"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/armon/go-radix"
	"github.com/maruel/panicparse/stack"
)

type nullWriter struct{}

func (w nullWriter) Write(buf []byte) (int, error) {
	return len(buf), nil
}

type config struct {
	server        string
	project       string
	environment   string
	depth         int
	includeVendor bool
	packages      []string
	watched       *radix.Tree
	cache         *sampleCache
}

func (c *config) shouldSample(fn string) bool {
	if _, _, present := c.watched.LongestPrefix(fn); !present {
		return false
	}

	if c.includeVendor {
		return true
	}

	return !strings.Contains(fn, "/vendor/")
}

func (c *config) sampleLoop(ctx context.Context) {
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
			for _, call := range g.Signature.Stack.Calls {
				if c.shouldSample(call.Func.Raw) {
					if err := c.cache.recordSample(call.Func.Raw); err != nil {
						log.Println("Error recording Mindsight sample:", err)
					}
					break
				}
			}
		}
	}
}

func (c *config) checkOptions() error {
	if c.server == "" {
		return errors.New("OptionAgentURL is required")
	}

	if c.project == "" {
		return errors.New("OptionProject is required")
	}

	if len(c.packages) == 0 {
		return errors.New("At least 1 package must be watched (OptionWatchPackage)")
	}

	return nil
}

type option func(*config)

func OptionWatchPackage(pkg string) option {
	return func(c *config) {
		c.packages = append(c.packages, pkg)
	}
}

func OptionAgentURL(url string) option {
	return func(c *config) {
		c.server = url
	}
}

func OptionProject(projectName string) option {
	return func(c *config) {
		c.project = projectName
	}
}

func OptionEnvironment(environment string) option {
	return func(c *config) {
		c.environment = environment
	}
}

func OptionCacheDepth(depth int) option {
	return func(c *config) {
		c.depth = depth
	}
}

func OptionIncludeVendor() option {
	return func(c *config) {
		c.includeVendor = true
	}
}

func StartMindsightCollector(ctx context.Context, options ...option) error {
	cfg := new(config)
	for _, opt := range options {
		opt(cfg)
	}

	if err := cfg.checkOptions(); err != nil {
		return err
	}

	watched := make(map[string]interface{})

	for _, p := range cfg.packages {
		watched[p] = nil
	}

	cfg.watched = radix.NewFromMap(watched)
	cfg.cache = newSampleCache(cfg)

	go cfg.sampleLoop(ctx)
	return nil
}
