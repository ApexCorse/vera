# Vera, simple CAN DBC parser and C code generator

> ...per la realizzazione di una macchina _vera_!
> 
> _Peppe Blunda_, [here](https://youtu.be/CIxJ0DwPr0U?si=nAdNeoc6L_XYmDUR&t=37)

`vera` was born for the application of the CAN DBC standard format, allowing us to define complex networks, and at the same time automate the writing of the code to encode/decode information inside CAN frames.

## Getting Started

The project uses the [Go programming language](https://golang.org) to build a simple-to-use, cross-platform CLI which allows to generate code based on a CAN DBC file.
It uses the `text/template` package to generate code from template files.

`vera` is also available as a standalone library, which you can add to your project via the following command:

```bash
go get -u github.com/ApexCorse/vera
```

### Development

Before writing code for mantainance/fixes/improvements, make sure you have installed the Go compiler.
Here are some useful commands:

```bash
# Run the CLI
go run cmd/vera/main.go

# Build an executable
go build cmd/vera/main.go

# Run tests
go build ./...

# Format code
go fmt ./...
```

The `gentest` directory contains a few (too few) tests that verify that the generated code actually works.
To run these tests, make sure you have [Make](https://www.gnu.org/software/make/) and [GCC](https://gcc.gnu.org/) compiler installed on your machine. Then execute:

```bash
cd gentest
make
```

### Using the CLI

You can either:

- Install a prebuilt binary for your OS from the [releases page](https://github.com/ApexCorse/vera/releases/).
- Build from scratch.

Then write your own `config.dbc` file and you're good to go. Run `vera -h` to view all the possible options.
