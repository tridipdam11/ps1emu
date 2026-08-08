# ps1emu

`ps1emu` is a modular, high-performance PlayStation 1 (PSX) emulator core written in Go. It implements the Sony PlayStation hardware specifications including the R3000A MIPS CPU core, System Control Coprocessor (COP0), pipeline delay slot emulation, and the PlayStation memory bus architecture.

---

## Features

### 🧠 CPU Core (MIPS R3000A)
- **MIPS I ISA Support**: Comprehensive instruction decoding and execution engine covering arithmetic, logical, shift, comparison, jump, branch, multiply/divide, and memory primitives.
- **Register File**: 32 General-Purpose Registers (R0–R31, R0 fixed to zero), 32-bit HI/LO multiplication and division registers.
- **System Control Coprocessor (COP0)**: Implements critical system registers including `Status` (SR, Reg 12), `Cause` (Reg 13), `EPC` (Reg 14), and `BadVAddr` (Reg 8).
- **Accurate Pipeline Delay Emulation**:
  - **Load Delays**: Realistic load delay slots with queuing, committing, and cancellation mechanics.
  - **Branch Delays**: Precise branch delay slot processing for conditional and unconditional branches/jumps.
- **Exception Pipeline**: Exception dispatching for Address Errors (Load/Store), Syscalls, Breakpoints, Reserved Instructions, Coprocessor Unusable, and Arithmetic Overflow with support for boot vectors (`0xBFC00180`) and standard vectors (`0x80000080`).
- **Unaligned Memory Access**: Built-in support for unaligned load/store word instructions (`LWL`, `LWR`, `SWL`, `SWR`).

### 💾 Memory Subsystem & Bus
- **Main RAM**: 2 MiB system RAM mapped at physical `0x00000000`–`0x001FFFFF` and mirrored across `0x00200000`–`0x007FFFFF`.
- **Scratchpad RAM**: Fast 1 KiB Data Cache / Scratchpad memory mapped at `0x1F800000`.
- **BIOS ROM**: 512 KiB Boot ROM mapped at `0x1FC00000` (physical) and accessible via KSEG1 (`0xBFC00000`).
- **Virtual Memory Address Translation**: Automatic translation across KUSEG (`0x00000000`), KSEG0 (`0x80000000`), and KSEG1 (`0xA0000000`) address spaces.

### ⚙️ System Interconnect
- Unified system container ([`system/system.go`](file:///D:/vscode/Projects/ps1emu/system/system.go)) connecting CPU and Memory Bus.
- Step-based execution API returning cycle counts and instruction metadata per step.

---

## Project Structure

```
ps1emu/
├── cpu/                # R3000A MIPS I CPU & COP0 implementation
│   ├── cpu.go          # CPU structure, register management & state reset
│   ├── decode.go       # Opcode, funct, and register field decoding
│   ├── instructions.go  # MIPS I instruction execution functions
│   ├── cop0.go         # Coprocessor 0 register operations (MFC0, MTC0, RFE)
│   ├── exceptions.go   # Exception vectors, Cause, SR & BadVAddr logic
│   ├── load_delay.go   # Queued load delay slot pipeline logic
│   ├── step.go         # Fetch-decode-execute step cycle engine
│   └── cpu_test.go     # Comprehensive CPU core unit tests
├── memory/             # PlayStation Memory Bus & Address Translation
│   ├── bus.go          # Memory space mapping, RAM mirroring, BIOS loading
│   └── bus_test.go     # Memory bus and region unit tests
├── system/             # System integration layer
│   ├── system.go       # System initialization & step interface
│   └── system_test.go   # Integration tests for CPU and Bus interaction
├── go.mod              # Go module definition
└── README.md           # Project documentation
```

---

## Getting Started

### Prerequisites
- **Go**: Version `1.25.0` or higher installed.

### Installation

Clone the repository to your local machine:

```bash
git clone https://github.com/tridipdam11/ps1emu.git
cd ps1emu
```

---

## Usage Example

The following example demonstrates initializing the system, loading a 512 KiB PlayStation BIOS file, and stepping through execution cycles:

```go
package main

import (
	"fmt"
	"os"

	"ps1emu/system"
)

func main() {
	// Read PlayStation BIOS file (512 KiB dump, e.g., SCPH1001.BIN)
	biosData, err := os.ReadFile("SCPH1001.BIN")
	if err != nil {
		fmt.Printf("Failed to read BIOS file: %v\n", err)
		return
	}

	// Initialize system with BIOS loaded
	sys, err := system.NewWithBIOS(biosData)
	if err != nil {
		fmt.Printf("Failed to initialize PS1 system: %v\n", err)
		return
	}

	fmt.Printf("System initialized. CPU starting PC: 0x%08X\n", sys.CPU.PC)

	// Step CPU for 10 cycles
	for i := 0; i < 10; i++ {
		res := sys.CPU.StepResult()
		fmt.Printf("[%d] PC: 0x%08X | Instruction: 0x%08X | Cycles: %d\n",
			i+1, res.PC, res.Instruction, res.Cycles)
	}
}
```

---

## Testing

Run all package unit tests across CPU, Memory Bus, and System packages:

```bash
go test ./...
```

To run tests with verbose logging and code coverage:

```bash
go test -v -cover ./...
```

---

## Architecture Details

### Memory Map Overview

| Address Range | Size | Description |
| :--- | :--- | :--- |
| `0x00000000 - 0x001FFFFF` | 2 MiB | Main RAM |
| `0x00200000 - 0x007FFFFF` | 6 MiB | Main RAM Mirror |
| `0x1F800000 - 0x1F8003FF` | 1 KiB | Scratchpad RAM (D-Cache) |
| `0x1FC00000 - 0x1FC7FFFF` | 512 KiB | BIOS Boot ROM |
| `0x80000000 - 0x9FFFFFFF` | 512 MiB | KSEG0 (Cached Kernel Memory, Physical `0x00000000`) |
| `0xA0000000 - 0xBFFFFFFF` | 512 MiB | KSEG1 (Uncached Kernel Memory, Physical `0x00000000`) |

---

## Roadmap

Future developments for full PSX emulation:
- [ ] **GPU**: 2D/3D hardware renderer, VRAM buffer, and command decoder
- [ ] **DMA Controller**: 7-channel Direct Memory Access controller
- [ ] **Timers**: 3 hardware root counters
- [ ] **CD-ROM Subsystem**: Disc image reader (`.bin`/`.cue`), command processing, and IRQ delivery
- [ ] **SPU**: 24-channel Sound Processing Unit
- [ ] **Input / Joypad**: Controller and Memory Card interface

---

## License

This project is licensed under the GNU General Public License v3.0 - see the [`LICENSE`](file:///D:/vscode/Projects/ps1emu/LICENSE) file for details.

