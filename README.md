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
- MQTT topic mapping and optional alert metadata via standard DBC signal attributes, to integrate in MQTT networks and pipelines
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
| `types.go` | Shared types (`Config`, `Node`, `Endianness`, `SignalMetadata`) |
| `validator.go` | Validation of DBC content and signal metadata |
| `errors.go` | Error construction with line number context |

### Codegen Package Files

The `codegen/` package handles C code generation:

| File | Description |
|------|-------------|
| `vera.{h,c}.tmpl` | Source and header file templates (used via `text/template`) |
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

    // The generated maximum avoids frame-ID-specific sizing and heap allocation.
    vera_decoded_signal_t signals[VERA_MAX_SIGNALS_PER_FRAME];
    vera_decoding_result_t result = VERA_DECODING_RESULT(signals);
    vera_err_t err = vera_decode_can_frame(&frame, &result);

    if (err == vera_err_ok) {
        for (uint8_t i = 0; i < result.n_signals; i++) {
            printf("Signal: %s = %.2f %s\n",
                   result.decoded_signals[i].name,
                   result.decoded_signals[i].value,
                   result.decoded_signals[i].unit);
        }
    }

    return 0;
}
```

`VERA_MAX_SIGNALS_PER_FRAME` is generated from the DBC and lets one caller-owned
buffer decode any supported frame without heap allocation. To inspect a specific
frame before decoding, use `vera_frame_signal_count(frame_id, &count)`. Decoding
returns `vera_err_insufficient_capacity` without writing output when the buffer
is too small, and `vera_err_unknown_frame` for an unsupported frame ID.

### SDK-Specific Usage

With `espidf`:

```c
#include "vera_espidf.h"

void process_can_message(const twai_frame_t* esp_frame) {
    vera_decoded_signal_t signals[VERA_MAX_SIGNALS_PER_FRAME];
    vera_decoding_result_t result = VERA_DECODING_RESULT(signals);
    vera_err_t err = vera_decode_espidf_rx_frame(esp_frame, &result);

    if (err == vera_err_ok) {
        // Process decoded signals...
    }
}
```

With `stm32hal`:

```c
#include "vera_stm32hal.h"

void process_can_message(CAN_RxHeaderTypeDef* can_header, uint8_t* data) {
    vera_decoded_signal_t signals[VERA_MAX_SIGNALS_PER_FRAME];
    vera_decoding_result_t result = VERA_DECODING_RESULT(signals);
    vera_err_t err = vera_decode_stm32hal_rx_frame(can_header, data, &result);

    if (err == vera_err_ok) {
        // Process decoded signals...
    }
}
```

## DBC File Format

Vera expects DBC files with the following format:

```
BO_ <message_id> <message_name>: <dlc> <transmitter>
    SG_ <signal_name> : <start_bit>|<length>@<endianness><sign> (<factor>,<offset>) [<min>|<max>] "<unit>" <receivers>
BA_DEF_ SG_ "VeraMqttTopic" STRING ;
BA_ "VeraMqttTopic" SG_ <message_id> <signal_name> "<mqtt_topic>";
```

**Important notes:**
- Start bit and length are in **bits**, DLC is in **bytes**
- Receivers are parsed if present, but not used in code generation
- Only **little-endian** (endiananness `1`) is currently supported
- Signal metadata uses standard DBC `BA_DEF_ SG_` declarations and `BA_ ... SG_` assignments.
- Vera recognizes `VeraMqttTopic` (`STRING`), `VeraWarningLow`, `VeraWarningHigh`, `VeraCriticalLow`, `VeraCriticalHigh` (`FLOAT`), and `VeraStaleAfterMs` (`INT`). Every recognized assignment needs a preceding matching declaration.
- All metadata is optional. Thresholds are alert metadata, separate from a signal's physical `[min|max]` range. `VeraStaleAfterMs` must be greater than zero when present.
- The legacy `vera:mqtt-topic=` comment format is rejected.

### Example DBC File

```
BO_ 123 EngineSpeed: 6 Engine
    SG_ EngineSpeed : 0|32@1+ (0.1,0) [0|8000] "RPM" DriverGateway
    SG_ BatteryTemperature : 32|16@1+ (1,0) [0|8000] "ºC" DriverGateway
BA_DEF_ SG_ "VeraMqttTopic" STRING ;
BA_DEF_ SG_ "VeraWarningHigh" FLOAT -1000000 1000000;
BA_DEF_ SG_ "VeraCriticalHigh" FLOAT -1000000 1000000;
BA_DEF_ SG_ "VeraStaleAfterMs" INT 1 4294967295;
BA_ "VeraMqttTopic" SG_ 123 EngineSpeed "vehicle/engine/speed";
BA_ "VeraWarningHigh" SG_ 123 EngineSpeed 7500;
BA_ "VeraCriticalHigh" SG_ 123 EngineSpeed 7900;
BA_ "VeraStaleAfterMs" SG_ 123 EngineSpeed 500;
BA_ "VeraMqttTopic" SG_ 123 BatteryTemperature "vehicle/battery/temperature";
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
│   ├── vera.c.tmpl        # Source file template
│   ├── vera.h.tmpl        # Header file template
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
