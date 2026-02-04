# Vera

A lightweight DBC file parser and C code generator for CAN bus message decoding. Vera automates the conversion of CAN message definitions into efficient C code, providing both direct decoding support and native SDK integrations.

## Purpose

Vera exists to simplify the workflow of working with CAN networks by providing:

- **Generated API from DBC files**: Parse any CAN database (DBC) file and automatically generate typed C code for encoding/decoding messages
- **SDK-specific adapters**: Pre-built HAL integrations for popular embedded platforms (ESP-IDF, STM32 HAL, AutoDevKit)
- **Type safety**: Strongly-typed structures for messages and signals, eliminating manual bit manipulation
- **Minimal runtime**: Generated code is self-contained and has no external dependencies beyond the C standard library

The generated code provides:
- Signal decoding from raw CAN frames with automatic scaling/offset conversion
- Signal encoding for creating CAN frames
- MQTT topic mapping via TP_ instructions
- Validation for out-of-bounds values

## Installation

### Prerequisites

- Go 1.25.1 or later
- C compiler (for building binaries)

### Install

```bash
# Clone the repository
git clone https://github.com/ApexCorse/vera.git
cd vera

# Install the CLI tool
go install ./cmd/vera
```

You can also find pre-built binaries in the [releases](https://github.com/ApexCorse/vera/releases) page.

## Quick Start

```bash
# Generate C code from a DBC file
vera -f network.dbc ./output

# This creates:
# - output/vera.h     (header file with type definitions and function declarations)
# - output/vera.c     (source file with decoding/encoding implementations)
```

To use with a specific SDK:

```bash
# ESP-IDF integration
vera -f network.dbc -sdk espidf ./output

# STM32 HAL integration
vera -f network.dbc -sdk stm32hal ./output

# AutoDevKit integration
vera -f network.dbc -sdk autodevkit ./output
```

## Architecture

### Core Components

```
vera/
├── cmd/vera/              # CLI entry point
├── codegen/               # C code generation package
│   ├── codegen.go         # Generic C code generation logic
│   ├── templates.go       # Header and source templates
│   ├── espidf/            # ESP-IDF-specific adapter generation
│   ├── stm32hal/          # STM32 HAL-specific adapter generation
│   └── autodevkit/        # AutoDevKit-specific adapter generation
└── internal/              # Core parsing and validation (main package files below)
```

### Main Package Files

The main package (`vera/`) contains the core functionality:

| File | Description |
|------|-------------|
| `parser.go` | Parses DBC files and returns a `Config` structure |
| `message.go` | `Message` struct with validation and line parsing |
| `signal.go` | `Signal` struct with validation and detailed parsing |
| `types.go` | Shared types (`Config`, `Node`, `Endianness`, `SignalTopic`) |
| `validator.go` | Validation of DBC content (signal placement, duplicate topics, etc.) |
| `errors.go` | Error construction with line number context |

### Codegen Package Files

The `codegen/` package handles C code generation:

| File | Description |
|------|-------------|
| `templates.go` | Core C templates (types, helpers, decoding/encoding logic) |
| `codegen.go` | Function for generating standard `vera.h` and `vera.c` |
| `espidf/` | ESP-IDF HAL adapter (decodes ESP's native `twai_frame_t` type) |
| `stm32hal/` | STM32 HAL adapter (decodes `CAN_RxHeaderTypeDef` and `CAN_TxHeaderTypeDef`) |
| `autodevkit/` | AutoDevKit adapter (decodes `CANTxFrame` type) |

## Working with Vera

### Using the CLI

```bash
# Basic usage
vera [options] <build_path>

# Options
-f <file>         DBC file path (default: config.dbc)
-sdk <sdk>        Target SDK: espidf, stm32hal, autodevkit
-v                Print version (from VERA_VERSION env var)
```

### Writing Code with Generated Headers

```c
#include "vera.h"

int main() {
    // Initialize a CAN frame with data
    vera_can_rx_frame_t frame = {
        .id = 0x7B,
        .dlc = 6,
        .data = {0x10, 0x20, 0x30, 0x40, 0x50, 0x60}
    };

    // Decode the message
    vera_decoding_result_t result = {0};
    vera_err_t err = vera_decode_can_frame(&frame, &result);

    if (err == vera_err_ok) {
        for (uint8_t i = 0; i < result.n_signals; i++) {
            printf("Signal: %s = %.2f %s\n",
                   result.decoded_signals[i].name,
                   result.decoded_signals[i].value,
                   result.decoded_signals[i].unit);
        }
        free(result.decoded_signals);
    }

    return 0;
}
```

### SDK-Specific Usage

With `espidf`:

```c
#include "vera_espidf.h"

void process_can_message(const twai_frame_t* esp_frame) {
    vera_decoding_result_t result = {0};
    vera_err_t err = vera_decode_espidf_rx_frame(esp_frame, &result);

    if (err == vera_err_ok) {
        // Process decoded signals...
        free(result.decoded_signals);
    }
}
```

With `stm32hal`:

```c
#include "vera_stm32hal.h"

void process_can_message(CAN_RxHeaderTypeDef* can_header, uint8_t* data) {
    vera_decoding_result_t result = {0};
    vera_err_t err = vera_decode_stm32hal_rx_frame(can_header, data, &result);

    if (err == vera_err_ok) {
        // Process decoded signals...
        free(result.decoded_signals);
    }
}
```

## Creating New HAL Extensions

Vera's HAL extension system uses a small, well-defined interface. To create support for a new SDK, implement the two functions in a new package under `codegen/`:

### Required Functions

```go
package yourhal

import (
    "io"
    "github.com/ApexCorse/vera"
)

// GenerateHeader writes the HAL-specific header file content
func GenerateHeader(w io.Writer, config *vera.Config) error

// GenerateSource writes the HAL-specific source file content
func GenerateSource(w io.Writer, config *vera.Config) error
```

### Example: Minimal HAL Implementation

```go
package yourhal

import (
    "fmt"
    "io"
    "github.com/ApexCorse/vera"
)

const (
    headerFile = `#ifndef VERA_YOURHAL_H
#define VERA_YOURHAL_H

#include "vera.h"
#include "your_hal_header.h"

// Declare your SDK-specific decoding function
// This function should convert your SDK's CAN frame format
// to the common `vera_can_rx_frame_t` structure
vera_err_t yourhal_decode_rx_frame(
    YourSDKFrameType* frame,
    vera_decoding_result_t* result
);

%s // Include encoding function declarations

#endif // VERA_YOURHAL_H`
)

const (
    sourceFile = `#include "vera_yourhal.h"

// Your SDK-specific decoding implementation
// 1. Convert input frame to vera_can_rx_frame_t
// 2. Call vera_decode_can_frame()
yourhal_decode_rx_frame(YourSDKFrameType* frame, vera_decoding_result_t* result) {
    vera_can_rx_frame_t vera_frame = {
        .id = convert_id(frame),        // Convert SDK ID format
        .dlc = frame->dlc,
        // Copy data bytes...
    };

    return vera_decode_can_frame(&vera_frame, result);
}

%s // Include encoding function definitions
`)
)

func GenerateHeader(w io.Writer, config *vera.Config) error {
    s := fmt.Sprintf(headerFile, generateEncodingDecls(config))
    if _, err := w.Write([]byte(s)); err != nil {
        return err
    }
    return nil
}

func GenerateSource(w io.Writer, config *vera.Config) error {
    s := fmt.Sprintf(sourceFile, generateEncodingDefs(config))
    if _, err := w.Write([]byte(s)); err != nil {
        return err
    }
    return nil
}

func generateEncodingDecls(config *vera.Config) string {
    var b strings.Builder
    for i, m := range config.Messages {
        b.WriteString(fmt.Sprintf("vera_err_t yourhal_encode_%s(\n", m.Name))
        b.WriteString("\tYourSDKTxFrame* frame,\n")
        for j, s := range m.Signals {
            b.WriteString(fmt.Sprintf("\tuint64_t %s", s.Name))
            if j < len(m.Signals)-1 {
                b.WriteString(",\n")
            } else {
                b.WriteString("\n")
            }
        }
        b.WriteString(");\n")
        if i < len(config.Messages)-1 {
            b.WriteString("\n")
        }
    }
    return b.String()
}

func generateEncodingDefs(config *vera.Config) string {
    var b strings.Builder
    for i, m := range config.Messages {
        b.WriteString(fmt.Sprintf("yourhal_encode_%s(", m.Name))
        b.WriteString("\tYourSDKTxFrame* frame,\n")
        for j, s := range m.Signals {
            b.WriteString(fmt.Sprintf("\tuint64_t %s", s.Name))
            if j < len(m.Signals)-1 {
                b.WriteString(",\n")
            } else {
                b.WriteString("\n")
            }
        }
        b.WriteString(") {\n")
        b.WriteString("\tmemset(frame->data, 0, 8);\n")
        b.WriteString(fmt.Sprintf("\tframe->id = 0x%X;\n", m.ID))
        b.WriteString(fmt.Sprintf("\tframe->dlc = %d;\n", m.DLC))
        // Insert each signal into the frame data...
        b.WriteString("\treturn vera_err_ok;\n")
        b.WriteString("}\n")
        if i < len(config.Messages)-1 {
            b.WriteString("\n")
        }
    }
    return b.String()
}
```

### Adding the SDK to the CLI

Edit `cmd/vera/main.go` to add the new SDK option:

```go
// Import your new package
import (
    "github.com/ApexCorse/vera/codegen/yourhal"
    // ... other imports
)

// In the main switch, add your case:
switch *sdk {
case "yourhal":
    yourhalGeneration(buildPath, config)
// ... other cases
case "":
default:
    fmt.Printf("fatal: sdk '%s' not supported\n", *sdk)
    os.Exit(1)
}
```

### SDK-Specific Implementation Notes

When implementing `GenerateHeader` and `GenerateSource` for a new SDK:

1. **Header file**: Include your SDK's CAN headers, declare your SDK-specific decode function, include standard `vera.h`

2. **Source file**: Implement the SDK-specific decode wrapper (converts SDK frame → `vera_can_rx_frame_t`), include helper decoding/encoding functions

3. **Decoding conversion**: The adapter should convert the SDK's frame format to `vera_can_rx_frame_t` and call `vera_decode_can_frame()` for the actual decoding logic

4. **Encoding**: Generate encode functions that populate the SDK's frame format with signal values

## DBC File Format

Vera expects DBC files with the following format:

```
BO_ <message_id> <message_name>: <dlc> <transmitter>
    SG_ <signal_name> : <start_bit>|<length>@<endianness><sign>(<integer_figures>,<decimal_figures>) (<factor>,<offset>) [<min>|<max>] "<unit>" <receivers>
TP_ <signal_name> <mqtt_topic>
```

**Important notes:**
- Start bit and length are in **bits**, DLC is in **bytes**
- Receivers are parsed but not used in code generation
- Only **little-endian** (endiananness `1`) is currently supported
- TP_ instructions are placed at the same level as BO_ instructions (not indented)

### Example DBC File

```
BO_ 123 EngineSpeed: 6 Engine
    SG_ EngineSpeed : 0|32@1+ (0.1,0) [0|8000] "RPM" DriverGateway
    SG_ BatteryTemperature : 32|16@1+(12,4) (1,0) [0|8000] "ºC" DriverGateway
TP_ EngineSpeed vehicle/engine/speed
TP_ BatteryTemperature vehicle/battery/temperature
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test file
go test ./parser_test.go

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Building from Source

```bash
# Build the CLI tool
go build ./cmd/vera

# Install for local use
go install ./cmd/vera
```

### Creating Releases

Releases are automatically created when you push a git tag:

```bash
# Create and push a version tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

The GitHub Actions workflow handles:
- Building binaries for Linux (amd64), Windows (amd64), macOS Intel (amd64), and macOS Apple Silicon (arm64)
- Creating a GitHub release with auto-generated notes
- Uploading platform binaries as release assets

### Project Structure Reference

```
.
├── cmd/vera/              # CLI entry point (main.go)
├── codegen/               # C code generation
│   ├── codegen.go         # Generic code generation
│   ├── templates.go       # Core C templates
│   ├── espidf/            # ESP-IDF HAL adapter
│   ├── stm32hal/          # STM32 HAL adapter
│   └── autodevkit/        # AutoDevKit adapter
├── gentest/               # Test infrastructure
│   ├── CMakeLists.txt     # CMake build config
│   ├── config-test.dbc    # Test DBC file
│   ├── test.c             # Test application
│   ├── test.sh            # Test runner script
│   └── unity/             # Unity test framework
├── vera/                  # Main package (core functionality)
│   ├── parser.go          # DBC parser
│   ├── message.go         # Message parsing/validation
│   ├── signal.go          # Signal parsing/validation
│   ├── types.go           # Shared types
│   ├── validator.go       # DBC validation
│   ├── errors.go          # Error types
│   ├── message_test.go    # Message tests
│   ├── parser_test.go     # Parser tests
│   ├── signal_test.go     # Signal tests
│   └── validator_test.go  # Validator tests
├── go.mod
├── go.sum
└── README.md
```

## Contributing

When contributing to Vera:

1. Write tests for new features or bug fixes
2. Update documentation as needed
3. Follow the existing code style
4. All tests should pass before submitting a pull request

## License

This project is part of the ApexCorse organization.
