# portkill

`portkill` is a small Linux command-line tool that stops the process listening
on a TCP port.

It uses `ss` to find the listener, shows the process and asks for confirmation.
Before sending `SIGTERM`, it checks the port again so it does not accidentally
signal a new process that reused the same port.

## Usage

```text
portkill
portkill help
portkill list
portkill <port> [--force]
```

Examples:

```bash
portkill                  # select a listener interactively
portkill list             # only show the list
portkill 3000
portkill 3000 --force
```

Running `portkill` with no arguments lists the listening TCP ports that expose a
PID. Enter the number beside a process to select it, then confirm with `y`.
Use `portkill list` when you only want to view the list.

Press `y` to terminate the process. Enter, `n`, or end-of-input cancels the
operation. Use `--force` (or `-f`) to skip the prompt.

## Install

The tool requires Linux, Go 1.23 or later, and `ss` from the `iproute2` package.

```bash
go install github.com/MarufHossain14/golang-projects/portkill@latest
```

To build from this directory instead:

```bash
go install .
```

Run the tests with:

```bash
go test ./...
```
