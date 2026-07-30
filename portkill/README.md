# portkill

portkill is a command line project in Go. The main goal is to give it a port number like 3000 and then it will find which process is using that port and stop it.

the project takes the arguments from the user, checks if the port and options
are valid and finds the process listening on a TCP port. It shows the process
and asks the user before stopping it.

when the program starts it goes to `main.go`. This gets the arguments from the
terminal, checks the options and prints the result.

the Linux process code is in `process.go`. It finds the PID, reads the process
information and sends the signal to stop it.

the port has to be a number from 1 to 65535 because that is the valid range for
network ports. It also checks things like giving two ports, using an unknown
option, or using options together that do not make sense.

to find the process on Linux it uses a command called `ss`. This shows the TCP
ports that are waiting for connections and the PID using each port. After
getting the PID, it reads `/proc/<pid>/cmdline` to get the full command if it
is available.

after showing the process it asks `Kill this process? (Y/n)`. If the answer is
yes, Go sends a SIGTERM signal to the PID. SIGTERM asks the process to shut
down cleanly.

`--dry-run` only shows what would be stopped and never sends the signal.
`--force` skips the question and sends the signal right away.

`--list` runs ss without one specific port and shows all the listening TCP
ports in a table. The table is made with Go's tabwriter so another package is
not needed.

`--json` uses Go's json package to print the process information in JSON. This
is useful if the output needs to be read by a script instead of a person.

for example:

```bash
go run . 3000
```

this checks port 3000 and shows the process name, PID and command if something
is listening there. Then it asks before stopping it.

```bash
go run . 3000 --dry-run
go run . 3000 --force
go run . --list
go run . --list --json
```

```bash
go run . 70000
```

this gives an error because the port is too high.

some things I learned in this part are how command line arguments work, how to
keep values in a struct, how to convert text into a number, how to return
errors, how to run a Linux command from Go, how to read process information
from `/proc`, how Linux signals work, how to make a table and JSON output and
how to test different inputs.

to see the help:

```bash
go run . --help
```

to use the project on ubuntu or WSL it needs Go and the `ss` command. ss is
usually already installed, but it can be installed with:

```bash
sudo apt update
sudo apt install iproute2
```

after downloading the project, go into the portkill folder and install it:

```bash
go install .
```

the Go bin folder has to be in PATH so the command can be used from any folder:

```bash
export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"
```

then it can be used like this:

```bash
portkill 3000
portkill 3000 --dry-run
portkill 3000 --force
portkill --list
portkill --list --json
portkill list
portkill list --json
portkill --help
portkill --version
```

there are also shorter versions of the options:

```bash
portkill 3000 -d       # dry run
portkill 3000 -f       # force
portkill -l            # list
portkill -l -j         # list as json
```

if the commands are forgotten, `portkill --help` shows them again. While the
program is asking for confirmation, `y` means yes, `n` means no and Ctrl+C
cancels the command.

to run the tests:

```bash
go test ./...
```

the project only uses the Go standard library.
