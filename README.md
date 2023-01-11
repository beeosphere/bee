# Bee

The bee module is the core of BeeOS bee agents. It connects bees to a hive and acts as the proxy between the hive and the concrete bee implementation.

## Installation

Use the go CLI to download the module.

```bash
go get github.com/beeosphere/bee
```

## Usage

```go
package main

import (
	"github.com/beeosphere/bee/core"
	"github.com/beeosphere/bee/host"
)

func main() {
	options := host.GetOptions()

	provider := func(channelType string) core.Channel {
        // Create new channel instances depending on the requested type
		return NewChannel()
	}
	core.NewBee(options, provider).Buzz()
}
```
