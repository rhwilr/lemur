package compiler

import (
	"fmt"
	"testing"

	"github.com/rhwilr/monkey/code"
)

type bytecodeTestCase struct {
	input    string
	expected []byte
}

func TestWrite(t *testing.T) {
	tests := []bytecodeTestCase{
		{
			input: "1 + 2",
			expected: []byte{
				2,                         // 2 constants
				0, 0, 0, 0, 0, 0, 0, 0, 1, // const 1
				0, 0, 0, 0, 0, 0, 0, 0, 2, // const 2
				byte(code.OpConstant), 0, 0, // opcode 0 (Load const) 0
				byte(code.OpConstant), 0, 1, // opcode 0 (Load const) 1
				byte(code.OpAdd), // opcode 1 (Add)
				byte(code.OpPop), // opcode 6 (Pop)
			},
		},
		{
			input: `
			let integer = 99;
			return integer;
			`,
			expected: []byte{
				1,                          // 1 constants
				0, 0, 0, 0, 0, 0, 0, 0, 99, // const int type 99
				byte(code.OpConstant), 0, 0, // opcode 0 (Load const) 0
				byte(code.OpSetGlobal), 0, 0, // opcode 11 (OpSetGlobal) 0
				byte(code.OpGetGlobal), 0, 0, // opcode 12 (OpGetGlobal) 0
				byte(code.OpReturnValue), // opcode 19 (OpReturnValue)
			},
		},
		{
			input: `
			let string = "ABC€";
			return string;
			`,
			expected: []byte{
				1,                          // 1 constants
				1, 													// const string type 
				0, 0, 0, 6,									// length of string
				65, 66, 67, 226, 130, 172,  // bytes for ABC€
				byte(code.OpConstant), 0, 0, // opcode 0 (Load const) 0
				byte(code.OpSetGlobal), 0, 0, // opcode 11 (OpSetGlobal) 0
				byte(code.OpGetGlobal), 0, 0, // opcode 12 (OpGetGlobal) 0
				byte(code.OpReturnValue), // opcode 19 (OpReturnValue)
			},
		},
		{
			input: `
			let f = fn(x) {return x};
			f(32);
			`,
			expected: []byte{
				0x02, 0x02, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x12, 0x00, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, nil, 0x1b, 0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x01, 0x17, 0x01, 0x06,
			},
		},
	}

	err := runBytecodeTests(t, tests)
	if err != nil {
		t.Fatalf("TestWrite failed: %s", err)
	}
}

/*
** Helpers
 */
func runBytecodeTests(t *testing.T, tests []bytecodeTestCase) error {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		code := compiler.Bytecode()

		print(code.Instructions.String())

		bytecode := code.Write()

		if len(tt.expected) != len(bytecode) {
			return fmt.Errorf("wrong instructions length.\nwant=%q\ngot =%q", tt.expected, bytecode)
		}

		for i, ins := range bytecode {
			if tt.expected[i] != ins {
				return fmt.Errorf("wrong byte at %d.\nwant=%q\ngot =%q", i, tt.expected[i], ins)
			}
		}
	}

	return nil
}
