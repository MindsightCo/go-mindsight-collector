package main

import (
	"context"
	"fmt"
	"time"

	"github.com/MindsightCo/go-mindsight-collector"
)

func main() {
	// only measure hotpaths for 100ms, for this example
	// to run indefinitely, use only context.Background(), or whichever context you like
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	collector.StartMindsightCollector(ctx, "http://localhost:8000/samples/", []string{"github.com/MindsightCo/go-mindsight-collector"})

	fmt.Scanln()
}
