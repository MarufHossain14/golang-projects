# portkill

`portkill` will be a small cross-platform CLI for finding and terminating the
process listening on a network port.

The project is being built in small, tested steps. The current milestone
provides the project structure plus help and version commands. Port lookup and
termination will be added in later milestones.

## Project structure

```text
portkill/
├── cmd/portkill/       # The small executable entry point
├── internal/cli/       # Shared command-line behavior
├── go.mod
└── README.md
```

The `internal` directory prevents these application-only packages from being
imported by unrelated projects.

## Run during development

```bash
go run ./cmd/portkill --help
go run ./cmd/portkill version
```

## Test

```bash
go test ./...
```

This milestone uses only the Go standard library.
