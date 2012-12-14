package main

import (
	"fmt"
	"log"
	"time"
)

const printDebug = false
const ramBegin = 0x4000

type cpu struct {
	memory  []byte
	romSize word

	// 8 bytes, each a bitfield of keys currently pressed.
	keyboard [8]byte
	shiftForce uint

	// Clock from boot, in cycles.
	clock uint64

	// Registers:
	a, i, r    byte
	f          flags
	bc, de, hl word

	// "prime" registers:
	ap            byte
	fp            flags
	bcp, dep, hlp word

	// 16-bit registers:
	sp, pc, ix, iy word

	// Interrupt flag.
	iff1 bool

	// Which IRQs should be handled.
	irqMask byte

	// Which IRQs have been requested by the hardware.
	irqLatch byte

	// Which NMIs should be handled.
	nmiMask byte

	// Which NMIs have been requested by the hardware.
	nmiLatch byte

	// Various I/O settings.
	modeImage byte

	// Root of instruction tree.
	root *instruction

	// Channel to get updates from.
	cpuUpdateCh chan<- cpuUpdate

	previousDumpTime   time.Time
	previousDumpClock  uint64
	previousYieldClock uint64
}

// Command to the CPU from the UI.
type cpuCommand struct {
	Cmd  string
	Data string
}

func (cpu *cpu) run(cpuCommandCh <-chan cpuCommand) {
	running := false
	shutdown := false

	timerCh := getTimerCh()

	handleCmd := func(msg cpuCommand) {
		switch msg.Cmd {
		case "boot":
			cpu.clearKeyboard()
			running = true
		case "shutdown":
			shutdown = true
		case "press", "release":
			cpu.keyEvent(msg.Data, msg.Cmd == "press")
		default:
			panic("Unknown CPU command " + msg.Cmd)
		}
	}

	for !shutdown {
		if running {
			select {
			case msg := <-cpuCommandCh:
				handleCmd(msg)
			case <-timerCh:
				cpu.timerInterrupt(true)
			default:
				cpu.step()
			}
		} else {
			handleCmd(<-cpuCommandCh)
		}
	}

	log.Print("CPU shut down")

	// No more updates.
	close(cpu.cpuUpdateCh)
}

func (cpu *cpu) fetchByte() byte {
	value := cpu.readMem(cpu.pc)
	cpu.pc++
	return value
}

func (cpu *cpu) fetchWord() (w word) {
	// Little endian.
	w.setL(cpu.fetchByte())
	w.setH(cpu.fetchByte())

	return
}

func (cpu *cpu) writeMem(addr word, b byte) {
	// xtrs:trs_memory.c
	// Check ROM writing. Harmless in real life, but may indicate a bug here.
	if addr < cpu.romSize {
		// ROM.
		panic(fmt.Sprintf("Tried to write %02X to ROM at %04X", b, addr))
	} else if addr >= ramBegin {
		// RAM.
		cpu.memory[addr] = b
	} else if addr >= screenBegin && addr < screenEnd {
		// Screen.
		cpu.memory[addr] = b
		cpu.cpuUpdateCh <- cpuUpdate{Cmd: "poke", Addr: int(addr), Data: int(b)}
	} else if addr == 0x37E8 {
		// Printer. Ignore, but could print ASCII byte to display.
	} else {
		// Ignore write.
	}
}

func (cpu *cpu) writeMemWord(addr word, w word) {
	// Little endian.
	cpu.writeMem(addr, w.l())
	cpu.writeMem(addr+1, w.h())
}

func (cpu *cpu) readMem(addr word) (b byte) {
	// Memory-mapped I/O.
	// http://www.trs-80.com/trs80-zaps-internals.htm#memmapio
	// xtrs:trs_memory.c
	if addr >= ramBegin {
		// RAM.
		b = cpu.memory[addr]
	} else if addr == 0x37E8 {
		// Printer. 0x30 = Printer selected, ready, with paper, not busy.
		b = 0x30
	} else if addr < cpu.romSize {
		// ROM.
		b = cpu.memory[addr]
	} else if addr >= screenBegin && addr < screenEnd {
		// Screen.
		b = cpu.memory[addr]
	} else if addr >= keyboardBegin && addr < keyboardEnd {
		// Keyboard.
		b = cpu.readKeyboard(addr)
	} else {
		// Unmapped memory.
		b = 0xFF
	}

	return
}

func (cpu *cpu) readMemWord(addr word) (w word) {
	w.setL(cpu.readMem(addr))
	w.setH(cpu.readMem(addr + 1))

	return
}

func (cpu *cpu) pushByte(b byte) {
	cpu.sp--
	cpu.writeMem(cpu.sp, b)
}

func (cpu *cpu) pushWord(w word) {
	cpu.pushByte(w.h())
	cpu.pushByte(w.l())
}

func (cpu *cpu) popByte() byte {
	cpu.sp++
	return cpu.readMem(cpu.sp - 1)
}

func (cpu *cpu) popWord() word {
	var w word

	w.setL(cpu.popByte())
	w.setH(cpu.popByte())

	return w
}

func (cpu *cpu) log(s string) {
	if printDebug {
		fmt.Print(s)
	}
}

func (cpu *cpu) logf(format string, arg ...interface{}) {
	if printDebug {
		fmt.Printf(format, arg...)
	}
}

func (cpu *cpu) logln() {
	if printDebug {
		fmt.Println()
	}
}
