# portkill

portkill is a command line project in Go. The main goal is to give it a port number like 3000 and then it will find which process is using that port and stop it.

right now the full project is not done yet. So far it can take the arguments
from the user, show the help and version, and check if the port and options are
valid.

when the program starts it goes to `cmd/portkill/main.go`. This gets the
arguments from the terminal and sends them to the cli package.

in the cli package there is an Options struct. This is where we keep the port
number and options like force, dry-run, list and json together. Then
`ParseOptions` goes through the arguments one by one and fills the struct.

the port has to be a number from 1 to 65535 because that is the valid range for
network ports. It also checks things like giving two ports, using an unknown
option, or using options together that do not make sense.

for example:

```bash
go run ./cmd/portkill 3000
```

this accepts 3000 as a valid port. It does not find the process yet because
that will be added in the next part.

```bash
go run ./cmd/portkill 70000
```

this gives an error because the port is too high.

some things I learned in this part are how command line arguments work, how to
keep values in a struct, how to convert text into a number, how to return
errors and how to test different inputs.

to see the help:

```bash
go run ./cmd/portkill --help
```

to run the tests:

```bash
go test ./...
```

for now the project only uses the Go standard library.
