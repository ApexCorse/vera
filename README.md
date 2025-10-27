# Vera

A lightweight DBC file parser and C code generator for CAN bus message decoding.

## Overview

Vera is a code generation tool that parses DBC (Database Container) files and generates C source code for decoding CAN (Controller Area Network) bus messages. It simplifies the process of implementing CAN message decoders in embedded systems by automating the conversion of message definitions into efficient C code.

## Features

- **DBC Parsing**: Reads and parses DBC files containing CAN message and signal definitions
- **C Code Generation**: Generates header (.h) and source (.c) files with decoding functions
- **Type Safety**: Generates strongly-typed structures for messages and signals
- **MQTT Topic Mapping**: Supports defining MQTT topics per signal using TP_ instruction
- **Signal Decoding**: Handles various signal properties including:
  - Little-endian byte ordering (only little-endian is currently supported)
  - Signed/unsigned values
  - Scaling factors and offsets
  - Min/max validation
  - Units

## Installation

### Prerequisites

- Go 1.25.1 or later

### Build from Source

```bash
git clone https://github.com/ApexCorse/vera.git
cd vera
go build -o vera ./cmd
```

## Usage

```bash
vera [options] <build_path>
```

### Options

- `-f <file>`: Path to the DBC file (default: `config.dbc`)

### Arguments

- `<build_path>`: Directory where the generated C files will be created

### Example

```bash
# Generate C code from a DBC file
vera -f network.dbc ./output

# This creates:
# - output/vera.h (header file with type definitions)
# - output/vera.c (source file with decoding functions)
```

## DBC File Format

Vera expects DBC files with the following format:

```
BO_ <message_id> <message_name>: <dlc> <transmitter>
    SG_ <signal_name> : <start_byte>|<length>@<endianness><sign> (<factor>,<offset>) [<min>|<max>] "<unit>" <receivers>
    TP_ <signal_name> <mqtt_topic>
```

> **Note**: DLC (Data Length Code), Start Byte, and Length are all expressed in **bytes**.

> **Note**: Receivers are currently parsed but not used in code generation.

### Example DBC File

```
BO_ 123 EngineSpeed: 6 Engine
    SG_ EngineSpeed : 0|4@1+ (0.1,0) [0|8000] "RPM" DriverGateway
    TP_ EngineSpeed vehicle/engine/speed
    SG_ BatteryTemperature : 4|2@1+ (1,0) [0|8000] "ºC" DriverGateway
    TP_ BatteryTemperature vehicle/battery/temperature
```

## Generated Code

The generated C code provides:

- **Type definitions** for CAN frames, messages, and signals
- **Decoding functions** that extract and convert signal values from raw CAN data
- **Error handling** for allocation failures and out-of-bounds values

### Example Usage of Generated Code

```c
#include "vera.h"

vera_can_rx_frame_t frame = {
    .id = 0x7B,
    .dlc = 6,
    .data = {0x10, 0x20, 0x30, 0x40, 0x50, 0x60}
};

vera_decoding_result_t result;
vera_err_t err = vera_decode_can_frame(&frame, &result);

if (err == vera_err_ok) {
    for (int i = 0; i < result.n_signals; i++) {
        printf("Signal: %s = %.2f %s\n",
               result.decoded_signals[i].name,
               result.decoded_signals[i].value,
               result.decoded_signals[i].unit);
    }
    free(result.decoded_signals);
}
```

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
.
├── README.md
├── cmd/                  # Command-line interface
│   └── vera.go
├── gentest/              # Test files and examples
│   ├── CMakeLists.txt
│   ├── config-test.dbc
│   ├── test.c
│   ├── test.sh
│   └── unity/           # Unity test framework
├── go.mod
├── go.sum
└── internal/
    ├── codegen/         # C code generator
    │   ├── codegen.go
    │   └── templates.go
    └── parser/          # DBC file parser
        ├── errors.go
        ├── parser.go
        ├── parser_test.go
        ├── types.go
        ├── utils.go
        └── validator.go
```

## License

This project is part of the ApexCorse organization.
