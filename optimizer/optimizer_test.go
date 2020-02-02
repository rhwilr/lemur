package optimizer

import (
	"testing"

	"github.com/rhwilr/lemur/lexer"
	"github.com/rhwilr/lemur/ast"
	"github.com/rhwilr/lemur/parser"
)


type optimizerTestCase struct {
	input                string
	expected                string
}

func TestExpressionsWithoutOptimizations(t *testing.T) {
	tests := []optimizerTestCase{
		{
			input: `let one = 1;let two = 2;`,
			expected: `let one = 1;let two = 2;`,
		},
		{
			input: `let two = 2;`,
			expected: `let two = 2;`,
		},
		{
			input: `input += 7;`,
			expected: `input+=7`,
		},
		{
			input: `input++;`,
			expected: `input++`,
		},
		{
			input: `let one = 1; while (one < 99) {}`,
			expected: `let one = 1;while ((one < 99)) {}`,
		},
		{
			input: `true >= false`,
			expected: `(true >= false)`,
		},
	}

	runOptimizerTests(t, tests)
}

func TestSimpleArithmeticCalculations(t *testing.T) {
	tests := []optimizerTestCase{
		{
			input: `let input = 1 + 1;`,
			expected: `let input = 2;`,
		},
		{
			input: `let input = 1 * 6;`,
			expected: `let input = 6;`,
		},
		{
			input: `let input = 0 * 6;`,
			expected: `let input = 0;`,
		},
		{
			input: `let input = (1 * 6) + 2;`,
			expected: `let input = 8;`,
		},
		{
			input: `let input = 9 + 2 - 1;`,
			expected: `let input = 10;`,
		},
		{
			input: `let input = (6/2) + 2;`,
			expected: `let input = 5;`,
		},
	}

	runOptimizerTests(t, tests)
}

func TestStringConcatinations(t *testing.T) {
	tests := []optimizerTestCase{
		{
			input: `let input = "Helo " + "World";`,
			expected: `let input = "Helo World";`,
		},
		{
			input: `puts("Helo" + " " + "World!");`,
			expected: `puts("Helo World!")`,
		},
	}

	runOptimizerTests(t, tests)
}

func TestComparisonOptimizations(t *testing.T) {
	tests := []optimizerTestCase{
		{
			input: `let input = 1 == 1;`,
			expected: `let input = true;`,
		},
		{
			input: `let input = 1 == 6;`,
			expected: `let input = false;`,
		},
		{
			input: `let input = 1 < 2;`,
			expected: `let input = true;`,
		},
		{
			input: `let input = 9 >= 8;`,
			expected: `let input = true;`,
		},
		{
			input: `while (0 < 99) {}`,
			expected: `while (true) {}`,
		},
		{
			input: `true == false`,
			expected: `false`,
		},
		{
			input: `true != false`,
			expected: `true`,
		},
		{
			input: `true || false`,
			expected: `true`,
		},
		{
			input: `true && false`,
			expected: `false`,
		},
		{
			input: `1 || 0`,
			expected: `true`,
		},
		{
			input: `12 && 0`,
			expected: `false`,
		},
	}

	runOptimizerTests(t, tests)
}

/*
** Helpers
 */
func runOptimizerTests(t *testing.T, tests []optimizerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(t, tt.input)

		optimizer := New(program)
		optimized, err := optimizer.Optimize()

		if err != nil {
			t.Errorf("error while optimizing programm: %s", err)
			continue
		}

		if optimized.String() != tt.expected {
			t.Errorf("optimizer did not perform the expected optimization.\n got=%q \n want=%q", optimized.String(), tt.expected)
		}
	}
}


func parse(t *testing.T, input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	parsed := p.ParseProgram()

	errors := p.Errors()
	if len(errors) > 0 {
		t.Fatalf("parse error: %s", errors)
	}

	return parsed
}