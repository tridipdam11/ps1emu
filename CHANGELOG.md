# Changelog

All notable changes to the `ps1emu` project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-08-08

### Added
- **MIPS R3000A CPU Core**: Full MIPS I instruction decoder, execution pipeline, and arithmetic engine.
- **Coprocessor 0 (COP0)**: Implemented status, cause, exception Program Counter (EPC), and bad virtual address registers.
- **Pipeline Delay Slots**: Realistic queued load delay slots and branch delay slot execution mechanics.
- **Exception Pipeline**: Address errors, syscalls, breakpoints, reserved instructions, and coprocessor unusable dispatching.
- **Memory Subsystem**: 2 MiB Main RAM with physical mirroring, 1 KiB Scratchpad RAM, and KUSEG/KSEG0/KSEG1 translation.
- **BIOS Loader**: 512 KiB BIOS dump loading at physical address `0x1FC00000` / virtual address `0xBFC00000`.
- **System Integration Layer**: Unified step-based CPU/Bus harness and state reset.
- **Executable CLI**: Added `cmd/ps1emu/main.go` entry point.
- **CI/CD**: Configured GitHub Actions multi-platform workflow (`.github/workflows/ci.yml`).
