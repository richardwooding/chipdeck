package chip8

import "fmt"

// Disassemble renders one opcode as a mnemonic, shared by the debugger
// panel and the tests.
func Disassemble(op uint16) string {
	x := op >> 8 & 0xF
	y := op >> 4 & 0xF
	nn := byte(op)
	nnn := op & 0x0FFF

	switch op >> 12 {
	case 0x0:
		switch op {
		case 0x00E0:
			return "CLS"
		case 0x00EE:
			return "RET"
		default:
			return fmt.Sprintf("SYS %03X", nnn)
		}
	case 0x1:
		return fmt.Sprintf("JP %03X", nnn)
	case 0x2:
		return fmt.Sprintf("CALL %03X", nnn)
	case 0x3:
		return fmt.Sprintf("SE V%X, %02X", x, nn)
	case 0x4:
		return fmt.Sprintf("SNE V%X, %02X", x, nn)
	case 0x5:
		if op&0xF == 0 {
			return fmt.Sprintf("SE V%X, V%X", x, y)
		}
	case 0x6:
		return fmt.Sprintf("LD V%X, %02X", x, nn)
	case 0x7:
		return fmt.Sprintf("ADD V%X, %02X", x, nn)
	case 0x8:
		switch op & 0xF {
		case 0x0:
			return fmt.Sprintf("LD V%X, V%X", x, y)
		case 0x1:
			return fmt.Sprintf("OR V%X, V%X", x, y)
		case 0x2:
			return fmt.Sprintf("AND V%X, V%X", x, y)
		case 0x3:
			return fmt.Sprintf("XOR V%X, V%X", x, y)
		case 0x4:
			return fmt.Sprintf("ADD V%X, V%X", x, y)
		case 0x5:
			return fmt.Sprintf("SUB V%X, V%X", x, y)
		case 0x6:
			return fmt.Sprintf("SHR V%X {, V%X}", x, y)
		case 0x7:
			return fmt.Sprintf("SUBN V%X, V%X", x, y)
		case 0xE:
			return fmt.Sprintf("SHL V%X {, V%X}", x, y)
		}
	case 0x9:
		if op&0xF == 0 {
			return fmt.Sprintf("SNE V%X, V%X", x, y)
		}
	case 0xA:
		return fmt.Sprintf("LD I, %03X", nnn)
	case 0xB:
		return fmt.Sprintf("JP V0, %03X", nnn)
	case 0xC:
		return fmt.Sprintf("RND V%X, %02X", x, nn)
	case 0xD:
		return fmt.Sprintf("DRW V%X, V%X, %X", x, y, op&0xF)
	case 0xE:
		switch nn {
		case 0x9E:
			return fmt.Sprintf("SKP V%X", x)
		case 0xA1:
			return fmt.Sprintf("SKNP V%X", x)
		}
	case 0xF:
		switch nn {
		case 0x07:
			return fmt.Sprintf("LD V%X, DT", x)
		case 0x0A:
			return fmt.Sprintf("LD V%X, K", x)
		case 0x15:
			return fmt.Sprintf("LD DT, V%X", x)
		case 0x18:
			return fmt.Sprintf("LD ST, V%X", x)
		case 0x1E:
			return fmt.Sprintf("ADD I, V%X", x)
		case 0x29:
			return fmt.Sprintf("LD F, V%X", x)
		case 0x33:
			return fmt.Sprintf("LD B, V%X", x)
		case 0x55:
			return fmt.Sprintf("LD [I], V%X", x)
		case 0x65:
			return fmt.Sprintf("LD V%X, [I]", x)
		}
	}
	return fmt.Sprintf("DW %04X", op)
}
