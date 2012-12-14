package main

import (
	"log"
)

// http://www.trs-80.com/trs80-zaps-internals.htm#keyboard13
const keyboardBegin = 0x3800
const keyboardEnd = keyboardBegin + 256
const (
	shiftNeutral = iota
	shiftForceDown
	shiftForceUp
)

type keyInfo struct {
	byteIndex, bitNumber, shiftForce uint
}

var keyMap = map[string]keyInfo{
	// http://www.trs-80.com/trs80-zaps-internals.htm#keyboard13
	"@": {0, 0, shiftForceUp},

	"A": {0, 1, shiftForceDown},
	"B": {0, 2, shiftForceDown},
	"C": {0, 3, shiftForceDown},
	"D": {0, 4, shiftForceDown},
	"E": {0, 5, shiftForceDown},
	"F": {0, 6, shiftForceDown},
	"G": {0, 7, shiftForceDown},
	"H": {1, 0, shiftForceDown},
	"I": {1, 1, shiftForceDown},
	"J": {1, 2, shiftForceDown},
	"K": {1, 3, shiftForceDown},
	"L": {1, 4, shiftForceDown},
	"M": {1, 5, shiftForceDown},
	"N": {1, 6, shiftForceDown},
	"O": {1, 7, shiftForceDown},
	"P": {2, 0, shiftForceDown},
	"Q": {2, 1, shiftForceDown},
	"R": {2, 2, shiftForceDown},
	"S": {2, 3, shiftForceDown},
	"T": {2, 4, shiftForceDown},
	"U": {2, 5, shiftForceDown},
	"V": {2, 6, shiftForceDown},
	"W": {2, 7, shiftForceDown},
	"X": {3, 0, shiftForceDown},
	"Y": {3, 1, shiftForceDown},
	"Z": {3, 2, shiftForceDown},

	"a": {0, 1, shiftForceUp},
	"b": {0, 2, shiftForceUp},
	"c": {0, 3, shiftForceUp},
	"d": {0, 4, shiftForceUp},
	"e": {0, 5, shiftForceUp},
	"f": {0, 6, shiftForceUp},
	"g": {0, 7, shiftForceUp},
	"h": {1, 0, shiftForceUp},
	"i": {1, 1, shiftForceUp},
	"j": {1, 2, shiftForceUp},
	"k": {1, 3, shiftForceUp},
	"l": {1, 4, shiftForceUp},
	"m": {1, 5, shiftForceUp},
	"n": {1, 6, shiftForceUp},
	"o": {1, 7, shiftForceUp},
	"p": {2, 0, shiftForceUp},
	"q": {2, 1, shiftForceUp},
	"r": {2, 2, shiftForceUp},
	"s": {2, 3, shiftForceUp},
	"t": {2, 4, shiftForceUp},
	"u": {2, 5, shiftForceUp},
	"v": {2, 6, shiftForceUp},
	"w": {2, 7, shiftForceUp},
	"x": {3, 0, shiftForceUp},
	"y": {3, 1, shiftForceUp},
	"z": {3, 2, shiftForceUp},

	"0": {4, 0, shiftForceUp},
	"1": {4, 1, shiftForceUp},
	"2": {4, 2, shiftForceUp},
	"3": {4, 3, shiftForceUp},
	"4": {4, 4, shiftForceUp},
	"5": {4, 5, shiftForceUp},
	"6": {4, 6, shiftForceUp},
	"7": {4, 7, shiftForceUp},
	"8": {5, 0, shiftForceUp},
	"9": {5, 1, shiftForceUp},

	"`": {4, 0, shiftForceDown}, // Simulate Shift-0.
	"!": {4, 1, shiftForceDown},
	"\"": {4, 2, shiftForceDown},
	"#": {4, 3, shiftForceDown},
	"$": {4, 4, shiftForceDown},
	"%": {4, 5, shiftForceDown},
	"&": {4, 6, shiftForceDown},
	"'": {4, 7, shiftForceDown},
	"(": {5, 0, shiftForceDown},
	")": {5, 1, shiftForceDown},

	":": {5, 2, shiftForceUp},
	";": {5, 3, shiftForceUp},
	",": {5, 4, shiftForceUp},
	"-": {5, 5, shiftForceUp},
	".": {5, 6, shiftForceUp},
	"/": {5, 7, shiftForceUp},

	"*": {5, 2, shiftForceDown},
	"+": {5, 3, shiftForceDown},
	"<": {5, 4, shiftForceDown},
	"_": {5, 5, shiftForceDown},
	">": {5, 6, shiftForceDown},
	"?": {5, 7, shiftForceDown},

	"Enter": {6, 0, shiftNeutral},
	"Clear": {6, 1, shiftNeutral},
	"Break": {6, 2, shiftNeutral},
	"Up": {6, 3, shiftNeutral},
	"Down": {6, 4, shiftNeutral},
	"Left": {6, 5, shiftNeutral},
	"Right": {6, 6, shiftNeutral},
	" ": {6, 7, shiftNeutral},
	"Shift": {7, 0, shiftNeutral},
}

func (cpu *cpu) clearKeyboard() {
	for i := 0; i < len(cpu.keyboard); i++ {
		cpu.keyboard[i] = 0
	}
	cpu.shiftForce = shiftNeutral
}

func (cpu *cpu) readKeyboard(addr word) byte {
	addr -= keyboardBegin

	var b byte

	for i, keys := range cpu.keyboard {
		if addr&(1<<uint(i)) != 0 {
			if i == 7 {
				// Modify keys based on the shift force.
				switch cpu.shiftForce {
				case shiftNeutral:
					// Nothing.
				case shiftForceUp:
					// On the Model III the first two bits are left and right shift,
					// though we don't handle the right shift anywhere.
					keys &^= 0x03
				case shiftForceDown:
					keys |= 0x01
				}
			}

			b |= keys
		}
	}

	return b
}

func (cpu *cpu) keyEvent(key string, isPressed bool) {
	log.Printf("Key %s is %v\n", key, isPressed)
	keyInfo, ok := keyMap[key]
	if !ok {
		log.Printf("Unknown key \"%s\"", key)
		return
	}

	cpu.shiftForce = keyInfo.shiftForce
	bit := byte(1 << keyInfo.bitNumber)
	if isPressed {
		cpu.keyboard[keyInfo.byteIndex] |= bit
	} else {
		cpu.keyboard[keyInfo.byteIndex] &^= bit
	}
}
