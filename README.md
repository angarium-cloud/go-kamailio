# go-kamailio

[![Go Report Card](https://goreportcard.com/badge/github.com/voiplens/go-kamailio)](https://goreportcard.com/report/github.com/voiplens/go-kamailio)
![CI](https://github.com/voiplens/go-kamailio/actions/workflows/tests.yml/badge.svg)
[![GoDoc](https://pkg.go.dev/badge/github.com/voiplens/go-kamailio/v3)](https://pkg.go.dev/github.com/voiplens/go-kamailio/v3)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/voiplens/go-kamailio/blob/master/LICENSE)

Zero dependencies Go implementation of Kamailio BINRPC protocol for invoking RPC functions.

This library works with any Kamailio version.

`go.voiplens.io/kamailio` requires at least Go 1.21.

## Usage

```go
import binrpc "go.voiplens.io/kamailio/binrpc"
```

### Full Example

```go
package main

import (
	"fmt"
	"net"

	binrpc "go.voiplens.io/kamailio/binrpc"
)

func main() {
	// establish connection to Kamailio server
	conn, err := net.Dial("tcp", "localhost:2049")

	if err != nil {
		panic(err)
	}

	// WritePacket returns the cookie generated
	cookie, err := binrpc.WritePacket(conn, "tm.stats")

	// for commands that require args, add them as function args
	// like this: binrpc.WritePacket(conn, "stats.fetch", "all")

	if err != nil {
		panic(err)
	}

	// the cookie is passed again for verification
	// we receive records in response
	records, err := binrpc.ReadPacket(conn, cookie)

	if err != nil {
		panic(err)
	}

	// "tm.stats" returns one record that is a struct
	// and all items are int values
	items, _ := records[0].StructItems()

	for _, item := range items {
		value, _ := item.Value.Int()

		fmt.Printf("%s = %d\n",
			item.Key,
			value,
		)
	}
}
```

### Kamailio Config

The `ctl` module must be loaded:

```
loadmodule "ctl.so"
```

If you are using `kamcmd` (and you most probably are), the module is already loaded.

In order to connect remotely, you must listen on TCP or UDP (defaults to local unix socket):

```
modparam("ctl", "binrpc", "tcp:2049")
```

**WARNING**: This will open your Kamailio to the world. Make sure you have a firewall in place, or listen on an internal interface.

## Limits

For now, only int double string and structs are implemented. Other types will return an error.

## Contributing

Contributions are welcome.

## License

This library is distributed under the [MIT](https://github.com/voiplens/go-kamailio/blob/master/LICENSE) license.
