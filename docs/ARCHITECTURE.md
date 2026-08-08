# PlayStation 1 Core Architecture (`ps1emu`)

This document provides a technical specification of the hardware modules emulated in `ps1emu`.

---

## 🏛️ System Architecture Overview

```
                        +--------------------------+
                        |  CPU: MIPS R3000A Core   |
                        | (32 GPRs, HI/LO, PC/nPC) |
                        +------------+-------------+
                                     |
                        +------------+-------------+
                        | COP0 System Coprocessor   |
                        | (SR, Cause, EPC, BadVAddr)|
                        +------------+-------------+
                                     |
  +----------------------------------+----------------------------------+
  |                                  |                                  |
+-+-------------------+    +---------+-----------+    +-----------------+-+
| Memory Bus (bus.go) |    | Direct Memory Access|    | GPU / Display   |
| RAM (2 MiB)         |    | (DMA - 7 Channels)  |    | (VRAM / Command)|
| Scratchpad (1 KiB)  |    +---------------------+    +-------------------+
| BIOS ROM (512 KiB)  |
+---------------------+
```

---

## 🧠 CPU Core (`cpu/`)

- **Instruction Set**: MIPS I (32-bit big/little endian instruction encoding).
- **Core Registers**:
  - `R0`–`R31`: General-purpose registers (`R0` hardwired to `0`).
  - `HI`, `LO`: Multiplication product and division quotient/remainder registers.
  - `PC`, `NextPC`, `CurrentPC`: Execution control registers supporting branch delay slots.
- **Pipeline Delay Slots**:
  - **Load Delay Slot**: Executing an instruction immediately after a load cannot read the loaded register; load updates are queued until the following cycle.
  - **Branch Delay Slot**: The instruction immediately following a jump or branch is executed before taking the target address.

---

## 💾 Memory Subsystem (`memory/`)

### Address Mapping & Translation

| Address Region | Virtual Address Range | Physical Address | Mirroring / Description |
| :--- | :--- | :--- | :--- |
| **KUSEG** | `0x00000000 - 0x7FFFFFFF` | `0x00000000 - 0x1FFFFFFF` | User space / Direct mapped RAM |
| **KSEG0** | `0x80000000 - 0x9FFFFFFF` | `0x00000000 - 0x1FFFFFFF` | Cached Kernel space |
| **KSEG1** | `0xA0000000 - 0xBFFFFFFF` | `0x00000000 - 0x1FFFFFFF` | Uncached Kernel space (BIOS @ `0xBFC00000`) |
| **Main RAM** | `0x00000000 - 0x001FFFFF` | `0x00000000` | 2 MiB Physical RAM |
| **RAM Mirrors** | `0x00200000 - 0x007FFFFF` | `0x00000000` | 6 MiB Mirrored RAM regions |
| **Scratchpad** | `0x1F800000 - 0x1F8003FF` | `0x1F800000` | 1 KiB Fast D-Cache Scratchpad |
| **BIOS ROM** | `0x1FC00000 - 0x1FC7FFFF` | `0x1FC00000` | 512 KiB Read-Only Memory |

---

## 🔄 Roadmap Components

- [ ] **`gpu/`**: 2D/3D VRAM buffer, Command Buffer Processor, Polygon Renderer.
- [ ] **`dma/`**: 7-Channel DMA Controller (MDEC in/out, GPU, CD-ROM, SPU, OTC, RAM).
- [ ] **`spu/`**: Sound Processing Unit (24 ADPCM voices, Reverb buffer, Pitch modulation).
- [ ] **`cdrom/`**: Disc image reader (`.bin`/`.cue`), sector buffer, IRQ delivery.
- [ ] **`timers/`**: 3 Root Counters (Dotclock, H-Blank, System Clock).
