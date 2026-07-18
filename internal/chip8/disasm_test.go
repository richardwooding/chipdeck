package chip8

import "testing"

func TestDisassemble(t *testing.T) {
	cases := map[uint16]string{
		0x00E0: "CLS",
		0x00EE: "RET",
		0x0123: "SYS 123",
		0x1234: "JP 234",
		0x2456: "CALL 456",
		0x3A42: "SE VA, 42",
		0x4B10: "SNE VB, 10",
		0x5120: "SE V1, V2",
		0x6C99: "LD VC, 99",
		0x7D01: "ADD VD, 01",
		0x8120: "LD V1, V2",
		0x8121: "OR V1, V2",
		0x8122: "AND V1, V2",
		0x8123: "XOR V1, V2",
		0x8124: "ADD V1, V2",
		0x8125: "SUB V1, V2",
		0x8126: "SHR V1 {, V2}",
		0x8127: "SUBN V1, V2",
		0x812E: "SHL V1 {, V2}",
		0x9340: "SNE V3, V4",
		0xA678: "LD I, 678",
		0xB789: "JP V0, 789",
		0xC5FF: "RND V5, FF",
		0xD125: "DRW V1, V2, 5",
		0xE29E: "SKP V2",
		0xE2A1: "SKNP V2",
		0xF307: "LD V3, DT",
		0xF30A: "LD V3, K",
		0xF315: "LD DT, V3",
		0xF318: "LD ST, V3",
		0xF31E: "ADD I, V3",
		0xF329: "LD F, V3",
		0xF333: "LD B, V3",
		0xF355: "LD [I], V3",
		0xF365: "LD V3, [I]",
		0x5121: "DW 5121", // malformed
		0xF3FF: "DW F3FF",
	}
	for op, want := range cases {
		if got := Disassemble(op); got != want {
			t.Errorf("Disassemble(%04X) = %q, want %q", op, got, want)
		}
	}
}
