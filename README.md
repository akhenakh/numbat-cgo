# numbat-cgo

Go bindings for [Numbat](https://numbat.dev), a statically typed programming language for scientific computations with first-class physical units support.

## Installation

```bash
go get github.com/akhenakh/numbat-cgo
```

## Precompiled Binaries

The Rust library is precompiled as static libraries for common architectures and bundled in this repository. This enables `go get` to work out of the box without requiring the Rust toolchain.

### Supported Platforms

| OS      | Architecture |
|---------|-------------|
| Linux   | amd64       |
| Linux   | arm64       |
| Linux   | riscv64     |
| macOS   | arm64       |
| Windows | amd64       |

## Usage

```go
package main

import (
    "fmt"
    "log"

    numbat "github.com/akhenakh/numbat-cgo"
)

func main() {
    ctx := numbat.NewContext()
    defer ctx.Free()

    // Evaluate expressions
    result, err := ctx.Interpret("1 km / 50 s")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.StringOutput) // "20 m/s"
    fmt.Println(result.Value)          // 20.0
    fmt.Println(result.Unit)           // "m/s"

    // Set variables
    if err := ctx.SetVariable("speed", 100, "km/h"); err != nil {
        log.Fatal(err)
    }

    result, err = ctx.Interpret("speed to m/s")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.StringOutput) // "≈ 27.7778 m/s"
}
```

## Building from Source

If you need to build for an unsupported platform or modify the Rust code:

1. Install Rust toolchain
2. Build the static library: `cargo build --release`
3. Place the resulting `.a` file in the appropriate `lib/<os>_<arch>/` directory
4. `go build`

## License
Numbat is licensed MIT by David Peter
