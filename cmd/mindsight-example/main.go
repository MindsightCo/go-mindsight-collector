package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/MindsightCo/go-mindsight-collector"
)

func main() {
	// only measure hotpaths for 100ms, for this example
	// to run indefinitely, use only context.Background(), or whichever context you like
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	depth := 5
	if depthEnv := os.Getenv("CACHE_DEPTH"); depthEnv != "" {
		depth, _ = strconv.Atoi(depthEnv)
	}

	options := []collector.Option{
		collector.OptionAgentURL("http://localhost:8000/samples/"),
		collector.OptionCacheDepth(depth),
		collector.OptionWatchPackage("github.com/MindsightCo/go-mindsight-collector"),
	}

	if err := collector.StartMindsightCollector(ctx, options...); err != nil {
		log.Fatal(err)
	}

	fmt.Scanln()
}
