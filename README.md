# mygo

A Cargo-like build tool and package manager for Go projects.

## Features

### Project Management

- **`new`** — Scaffold a new Go project (binary or library) with sensible defaults
- **`init`** — Initialize `mygo.toml` in an existing project
- **`build`** — Compile with profiles (dev/release), cross-compilation, colored output
- **`run`** — Run the project with profile support
- **`test`** — Run tests with coverage, race detection, verbose output
- **`check`** — Fast compilation check (no binary output)
- **`clean`** — Remove build artifacts and caches
- **`fmt`** — Format Go source code
- **`vet`** — Run `go vet` with strict mode

### Dependency Management

- **`add`** — Add a dependency
- **`remove`** — Remove a dependency
- **`update`** — Update dependencies (with `--dry-run`)
- **`tidy`** — Clean up `go.mod`
- **`vendor`** — Vendor dependencies
- **`tree`** — Display dependency tree
- **`why`** — Show why a package is needed
- **`outdated`** — List outdated dependencies

### Advanced Tooling

- **`install`** / **`uninstall`** — Install / uninstall Go binaries
- **`doc`** — Serve documentation locally
- **`bench`** — Run benchmarks
- **`generate`** — Run `go generate`
- **`fix`** — Run `go fix`
- **`audit`** — Check for known vulnerabilities (`govulncheck`)
- **`workspace`** — Manage Go workspaces (`init`, `add`, `remove`, `sync`)
- **`env`** — Show Go environment variables
- **`config`** — Manage `mygo.toml`

## Installation

```bash
go install github.com/Ruixi-rebirth/mygo/cmd/mygo@latest
```

Or build from source:

```bash
git clone https://github.com/Ruixi-rebirth/mygo
cd mygo
CGO_ENABLED=0 go build -o mygo ./cmd/mygo/
```

## Quick Start

```bash
# Create a new project
mygo new --name myapp
cd myapp

# Build and run
mygo build
mygo run

# Add a dependency
mygo add github.com/gin-gonic/gin

# Run tests
mygo test -v --cover

# Check for outdated deps
mygo outdated

# Build for release
mygo build --release --target linux/amd64
```

## Configuration

`mygo.toml` (optional, at project root):

```toml
[project]
name = "myapp"
version = "0.1.0"

[build]
ldflags = "-s -w"
out-dir = "bin"

[profile.release]
ldflags = "-s -w"
env = { CGO_ENABLED = "0" }

[test]
timeout = "30s"
```

## License

MIT
