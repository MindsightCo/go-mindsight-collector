# Go Mindsight Collector

This utility can be plugged into your go application to collect vital data about your code's behavior so that Mindsight can help you write better code more safely, without significantly impacting your .

## Configuration

Before you get started, make sure you set up the [Mindsight Agent](https://github.com/MindsightCo/hotpath-agent), which is
required to send diagnostic data to Mindsight's backend for further analysis.

Let's assume for the example below that your agent will be listening at `http://localhost:8000`.

To start collecting data from your application, do the following your application (`main` is probably where you want to do this):

```
import (
  "context"
  "github.com/MindsightCo/go-mindsight-collector"
)

... // in a function, such as main:
ctx := context.Background()
collector.StartMindsightCollector(ctx,
    collector.OptionAgentURL("http://localhost:8000/samples/"),
    collector.OptionWatchPackage("github.com/you/your-package"),
    collector.OptionWatchPackage("github.com/you/other-package"))
```

The hotpaths for the packages specified via `OptionWatchPackage` will be measured periodically and reported to the Mindsight backend via the [Mindsight Agent](https://github.com/MindsightCo/hotpath-agent).

### Optional Configuration

You can control how frequently samples are sent to the Agent via `collector.OptionCacheDepth()`.

Feel free to provide your own [context](https://godoc.org/context) according to your needs. The collector will halt if the context receives a cancellation request (i.e. it respects `ctx.Done()`).
