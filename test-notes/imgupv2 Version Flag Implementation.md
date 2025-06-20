---
date: 2025-05-31 15:45:59
title: imgupv2 Version Flag Implementation
type: note
permalink: basic-memory/imgupv2-version-flag-implementation
tags:
  - imgupv2 cli version-flag golang cobra development
  - imgupv2
  - cli
  - golang
  - build
  - git
  - versioning
  - cobra
modified: 2025-06-01 07:57:45
---

# imgupv2 Version Flag Implementation

## Summary

Successfully implemented version flag support for the imgupv2 CLI tool, allowing users to check the version via `--version`/`-v` flags or the `version` subcommand.
- [feature] Version flag support via --version/-v flags and version subcommand #imgupv2 #cli (standard CLI pattern)

## Implementation Details

### Version Variables

Added three version variables to `cmd/imgup/main.go` that are populated at build time:

```go
var (
    version = "dev"
    commit  = "unknown"
    date    = "unknown"
)
```

- [implementation] Version variables populated at build time using ldflags #imgupv2 #golang #build (standard Go pattern)

### Command Support

1. **Version subcommand**: `imgup version`
2. **Version flag**: `imgup --version` or `imgup -v`

Both display:

```
imgupv2 version 0.2.1
  commit: fa4f2e6
  built:  2025-05-31T05:09:00Z
```

### Build Process

The key insight was that the build must be run from within the `cmd/imgup` directory for ldflags to work correctly:

```bash
cd cmd/imgup
go build -ldflags "-X 'main.version=0.2.1' -X 'main.commit=$(git rev-parse --short HEAD)' -X 'main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o ../../imgup .
```

- [learning] Build must run from cmd/imgup directory for ldflags to work correctly #imgupv2 #golang #build (package path resolution)

### Makefile Update

Updated the Makefile to automatically inject version information:

```makefile
build:
	cd cmd/imgup && go build -ldflags "-X 'main.version=$$(git describe --tags --always --dirty)' -X 'main.commit=$$(git rev-parse --short HEAD)' -X 'main.date=$$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o ../../imgup .
```

- [implementation] Makefile uses git describe --tags --always --dirty for meaningful version strings #imgupv2 #git #versioning (automatic versioning)

### Goreleaser Integration

The `.goreleaser.yaml` already had the correct ldflags configuration:

```yaml
ldflags:
  - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
```

## Technical Notes

### Cobra Version Flag Issue

- Cobra's automatic `--version` flag (via `Version` field) didn't work as expected
- Implemented manual flag handling in the root command's `RunE` function
- This approach is more reliable and gives us full control over the output format
- [issue] Cobra's automatic --version flag didn't work as expected, required manual implementation #imgupv2 #cobra (framework limitation)

### Build Location Matters

- Building from the project root with `go build -o imgup ./cmd/imgup` didn't properly inject the ldflags
- Building from within `cmd/imgup` directory resolved the issue
- This is likely due to how Go resolves package paths for ldflags
- [gotcha] Building from project root doesn't inject ldflags properly - must build from cmd/imgup #imgupv2 #golang #build (common mistake)

## Usage Examples

```bash
# Check version via flag
imgup --version
imgup -v

# Check version via subcommand
imgup version

# Build with version info
make build

# Build with custom version
cd cmd/imgup
go build -ldflags "-X 'main.version=1.2.3'" -o ../../imgup .
```

## Future Considerations

- The version is now properly set during releases via goreleaser
- Local development builds show "dev" unless explicitly set
  - [convention] Local development builds show "dev" version unless explicitly set #imgupv2 #versioning (development practice)
- The `git describe --tags --always --dirty` in the Makefile provides meaningful version strings even between releases
