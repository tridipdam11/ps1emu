# Contributing to ps1emu

Thank you for your interest in contributing to `ps1emu`! This project aims to build a modular, well-documented PlayStation 1 (PSX) emulator written in Go.

---

## 🛠️ Development Setup

### Prerequisites
- **Go**: Version `1.25.0` or later installed.
- **Git**: Installed and configured.

### Getting Started
1. **Fork & Clone** the repository:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ps1emu.git
   cd ps1emu
   ```
2. **Run existing tests** to verify your setup:
   ```bash
   go test -v ./...
   ```

---

## 📐 Project Conventions & Architecture

- **`cpu/`**: R3000A MIPS I CPU core, register sets, COP0 system coprocessor, delay slots, and exception vectors.
- **`memory/`**: 2 MiB RAM, 1 KiB Scratchpad, 512 KiB BIOS ROM address decoding and KSEG virtual address translation.
- **`system/`**: Top-level hardware bus connector and execution loop.
- **`cmd/ps1emu/`**: CLI entry point executable.

### Code Style
- Run `gofmt -s -w .` before committing code.
- Ensure all exported functions, types, and constants have descriptive Go doc comments.
- Keep hardware component emulation decoupled and deterministic.

---

## 🧪 Testing Guidelines

Every new feature, CPU opcode, or memory region change MUST include corresponding unit tests:
- CPU instruction tests belong in [`cpu/cpu_test.go`](file:///D:/vscode/Projects/ps1emu/cpu/cpu_test.go).
- Memory bus and mirroring tests belong in [`memory/bus_test.go`](file:///D:/vscode/Projects/ps1emu/memory/bus_test.go).
- System integration tests belong in [`system/system_test.go`](file:///D:/vscode/Projects/ps1emu/system/system_test.go).

Run the full test suite with coverage:
```bash
go test -v -cover ./...
```

---

## 📬 Submitting Pull Requests

1. Create a feature branch (`git checkout -b feature/gpu-vram`).
2. Commit changes with clean commit messages.
3. Ensure all unit tests pass (`go test ./...`).
4. Push your branch and open a Pull Request against `main`.
