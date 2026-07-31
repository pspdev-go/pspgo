# pspgo

[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen?style=flat-square)](/LICENSE)
[![Release](https://github.com/pspdev-go/pspgo/actions/workflows/release.yaml/badge.svg)](https://github.com/pspdev-go/pspgo/actions/workflows/release.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/pspdev-go/pspgo.svg)](https://pkg.go.dev/github.com/pspdev-go/pspgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/pspdev-go/pspgo)](https://goreportcard.com/report/github.com/pspdev-go/pspgo)
[![CI](https://github.com/pspdev-go/pspgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/pspdev-go/pspgo/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/pspdev-go/pspgo/graph/badge.svg?token=O62UTU7SHA)](https://codecov.io/gh/pspdev-go/pspgo)

`pspgo` is a build tool for PSP applications written with
[`pspsdk-go`](https://github.com/pspdev-go/pspsdk-go). It compiles Go code with
TinyGo, resolves the required PSPSDK libraries and bridge sources, links the
application with the PSP toolchain, and packages the result as `EBOOT.PBP`.

`pspgo` is designed specifically for `pspsdk-go` projects. It is not a
general-purpose Go compiler or a replacement for TinyGo and PSPSDK.

## Requirements

Before installing `pspgo`, prepare:

- [Go](https://go.dev/)
- Forked [TinyGo](https://github.com/pspdev-go/tinygo) with PSP support
- [PSPSDK](https://github.com/pspdev/pspdev)
- CMake
- a checkout of [`pspsdk-go`](https://github.com/pspdev-go/pspsdk-go)

Set `PSPDEV` to the PSPSDK installation directory and ensure the PSPSDK tools
are available:

```sh
export PSPDEV="$HOME/pspdev"
export PATH="$PSPDEV/bin:$PATH"
```

`pspgo` reads the Go version required by the project's `go.mod` and asks Go's
built-in toolchain manager for that version. A newer Go on `PATH` is therefore
safe. The required toolchain may be downloaded on the first run.

## Installation

### Build from source

Clone the repository and build the executable:

```sh
git clone https://github.com/pspdev-go/pspgo.git
cd pspgo
go build -o pspgo .
```

Move the resulting `pspgo` executable to a directory on `PATH`.

### Install with `go install`

```sh
go install github.com/pspdev-go/pspgo@latest
```

Make sure the Go binary directory, normally `$(go env GOPATH)/bin`, is on
`PATH`.

### Download from the releases page

Prebuilt executables are available on the
[GitHub Releases page](https://github.com/pspdev-go/pspgo/releases). Download
the archive for your operating system and architecture, extract it, and place
the `pspgo` executable on `PATH`.

## Quick start with `pspsdk-go`

From the root of a `pspsdk-go` checkout:

```sh
pspgo doctor
pspgo build ./example
```

`doctor` checks Go, TinyGo, PSPSDK, and CMake before building. The generated
application is:

```text
build/pspgo/cmake/EBOOT.PBP
```

For an application in a separate directory, point `pspgo` at the
`pspsdk-go` checkout:

```sh
export PSPGO_SDK=/path/to/pspsdk-go
cd /path/to/my-pspsdk-go-app

pspgo doctor
pspgo build .
```

The SDK is detected automatically when the project or one of its parent
directories contains `bridge/main.c`.

## Commands

```text
pspgo doctor                  check the build environment
pspgo env                     print the resolved configuration and tools
pspgo build [package]         build and package an EBOOT.PBP
pspgo build -v [package]      build while printing external commands
pspgo run [package]           build and launch the configured PPSSPP
pspgo test [package]          compile, link, and package test targets
pspgo clean                   remove build/pspgo
```

`pspgo test` currently verifies compilation, linking, and packaging. It does
not yet deploy tests to PPSSPP or collect runtime results.

## Configuration

Configuration can be supplied through environment variables:

| Variable          | Purpose                       |
| ----------------- | ----------------------------- |
| `PSPGO_SDK`       | `pspsdk-go` checkout          |
| `PSPGO_GO`        | bootstrap Go executable       |
| `PSPGO_TINYGO`    | TinyGo executable             |
| `PSPGO_PSP_CMAKE` | `psp-cmake` executable        |
| `PSPGO_PSP_NM`    | `psp-nm` executable           |
| `PSPGO_CMAKE`     | CMake executable              |
| `PSPGO_PPSSPP`    | PPSSPP executable             |
| `PSPDEV`          | PSPSDK installation directory |

A project can instead contain a `pspgo.toml`:

```toml
title = "My PSP Game"
output = "my-game"
sdk_root = "../pspsdk-go"
build_dir = "build/pspgo"
kernel_mode = false

# Optional tool overrides:
# go = "/path/to/go"
# tinygo = "/path/to/tinygo"
# psp_cmake = "/path/to/psp-cmake"
# psp_nm = "/path/to/psp-nm"
# ppsspp = "/path/to/PPSSPPSDL"
```

The PSP TinyGo target is embedded in `pspgo` and is materialized inside the
build directory. A project does not need its own `psp.json`. Set the optional
`target` key only when overriding the embedded target:

```toml
target = "/path/to/custom-psp.json"
```

## How it works

`pspgo` owns the build orchestration:

1. It selects the Go toolchain required by the project.
2. TinyGo compiles the selected package with the embedded PSP target.
3. `psp-nm` discovers undefined PSP symbols in the generated object.
4. `pspgo` selects the required `pspsdk-go` bridge sources and PSPSDK
   libraries.
5. PSP CMake and GCC link the application and package `EBOOT.PBP`.

`pspgo` invokes `cmake --build` and does not call Make directly. CMake may use
Make, Ninja, or another available build backend depending on its configured
generator.

`pspsdk-go` remains the source of truth for PSP API packages, startup code,
library requirement markers, and ABI adapters. `pspgo` supplies the compiler
target and turns those components into a reproducible build pipeline.

## Current limitations

- Runtime test deployment and PPSSPP result collection are not implemented.
- Builds are not yet content-addressed.
- TinyGo may retain checkout paths in package-name strings even when debug
  information is disabled.
- Calls that have not been verified against the PSP ABI may require an
  explicit C adapter supplied by `pspsdk-go`.
