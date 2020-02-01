package compiler

import (
	"testing"

	"github.com/rhwilr/lemur/code"
)


func TestConstantOptimizations(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"lemur"`,
			expectedConstants: []interface{}{"lemur"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `1 + 2`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `1 + 1`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"lemur" + "lemur"`,
			expectedConstants: []interface{}{"lemur"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}
