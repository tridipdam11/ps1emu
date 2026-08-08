package main

import (
	"flag"
	"fmt"
	"os"

	"ps1emu/system"
)

func main() {
	biosPath := flag.String("bios", "", "Path to PlayStation 1 BIOS dump file (512 KiB, e.g. SCPH1001.BIN)")
	steps := flag.Int("steps", 10, "Number of CPU instruction cycles to step through")
	flag.Parse()

	fmt.Println("==================================================")
	fmt.Println("  ps1emu - PlayStation 1 (PSX) Emulator Core")
	fmt.Println("==================================================")

	var sys *system.System
	var err error

	if *biosPath != "" {
		fmt.Printf("Loading BIOS file: %s\n", *biosPath)
		biosData, readErr := os.ReadFile(*biosPath)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to read BIOS file: %v\n", readErr)
			os.Exit(1)
		}

		sys, err = system.NewWithBIOS(biosData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize system with BIOS: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("BIOS successfully loaded into memory address 0x1FC00000.")
	} else {
		fmt.Println("No BIOS specified. Initializing empty system core...")
		sys = system.New()
	}

	fmt.Printf("Initial CPU State -> PC: 0x%08X | Status (COP0 R12): 0x%08X\n\n", sys.CPU.PC, sys.CPU.Cop0[12])

	fmt.Printf("Executing %d CPU step cycles:\n", *steps)
	for i := 1; i <= *steps; i++ {
		res := sys.CPU.StepResult()
		fmt.Printf("[%4d] PC: 0x%08X | Instruction: 0x%08X | Cycles: %d\n",
			i, res.PC, res.Instruction, res.Cycles)
	}

	fmt.Println("\nExecution finished successfully.")
}
