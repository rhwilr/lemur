package vm

import (
	"fmt"
	"testing"

	"github.com/rhwilr/lemur/ast"
	"github.com/rhwilr/lemur/compiler"
	"github.com/rhwilr/lemur/lexer"
	"github.com/rhwilr/lemur/object"
	"github.com/rhwilr/lemur/optimizer"
	"github.com/rhwilr/lemur/parser"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 <= 2", true},
		{"1 >= 2", false},
		{"1 <= 1", true},
		{"1 >= 1", true},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"!!0", false},
		{"!(if (false) { 5; })", true},
		{`"string" == "string"`, true},
		{`"string" == "String"`, false},
		{`"string" == "word"`, false},
		{`"string" != "string"`, false},
		{`"string" != "String"`, true},
		{`"string" != "word"`, true},
		{`"abc123" == "abc" + "123"`, true},
		{`"a" > "A"`, true},
		{`"a" < "A"`, false},
		{`"a" >= "a"`, true},
		{`"a" <= "z"`, true},
		{`"z" <= "z"`, true},
	}

	runVmTests(t, tests)
}

func TestLogicalExpressionsWithShortCircuit(t *testing.T) {
	tests := []vmTestCase{
		{`true && true`, true},
		{`true && false`, false},
		{`false && true`, false},
		{`true || false`, true},
		{`false || true`, true},
		{`1 || 0`, true},
		{`0 || 5`, true},
		{`6 && 5`, true},
		{`12 && 0`, false},
		{`if(true && true) { "a" }`, "a"},
		{`if(false && true) { "a" } else { "b" } `, "b"},
	}

	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"lemur"`, "lemur"},
		{`"le" + "mur"`, "lemur"},
		{`"le" + "mur" + "banana"`, "lemurbanana"},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 } ", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", Null},
		{"if (false) { 10 }", Null},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
		{"if (true && 1) { 10 }", 10},
		{"if (false && true) { 10 }", Null},
		{"if (false || true) { 10 }", 10},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
		{"let a = 5; a = 6;", 6},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}

	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			"{}", map[object.HashKey]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", Null},
		{"[1, 2, 3][99]", Null},
		{"[1][-1]", Null},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", Null},
		{"{}[0]", Null},
		{`"Hello"[0]`, "H"},
		{`"Hello"[1]`, "e"},
		{`"Hello"[1+1]`, "l"},
		{`"Hello"[100]`, Null},
		{`"Hello"[-1]`, Null},
	}

	runVmTests(t, tests)
}

func TestFunctionApplication(t *testing.T) {
	tests := []vmTestCase {
		{"let identity = function(x) { x; }; identity(5);", 5},
		{"let identity = function(x) { return x; }; identity(5);", 5},
		{"let double = function(x) { x * 2; }; double(5);", 10},
		{"let add = function(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = function(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},

		{"function identity (x) { x; }; identity(5);", 5},
		{"function identity (x) { return x; }; identity(5);", 5},
		{"function double (x) { x * 2; }; double(5);", 10},
		{"function add (x, y) { x + y; }; add(5, 5);", 10},
		{"function add (x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},

		{"function(x) { x; }(5)", 5},
		{"function(x = 5) { x; }()", 5},
		{"function(x, y = 5) { x + y; }(5)", 10},
		{"function(x, y = 5) { x + y; }(5, 10)", 15},
		{"function(x, b = false) { b; }(5)", false},
		{"function(x, b = false) { b; }(5, true)", true},
		{"function(x, b = false, y = 5) { x + y; }(5)", 10},
		{"function(x, b = false, y = 5) { x + y; }(5, true)", 10},
	}

	runVmTests(t, tests)
}
func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let fivePlusTen = function() { 5 + 10; };
			fivePlusTen();
			`,
			expected: 15,
		},
		{
			input: `
			let one = function() { 1; };
			let two = function() { 2; };
			one() + two()
			`,
			expected: 3,
		},
		{
			input: `
			let a = function() { 1 };
			let b = function() { a() + 1 };
			let c = function() { b() + 1 };
			c();
			`,
			expected: 3,
		},
		{
			input: `
			let a = function(x = 1) { x };
			let b = function(x = 2) { a() + x };
			let c = function(i = 3, j = 2) { b() + i - j };
			c();
			`,
			expected: 4,
		},
		{
			input: `
			let c = function(h, i = 3, j = 2) { h + i + j };
			c(1, 1, 1);
			`,
			expected: 3,
		},
	}

	runVmTests(t, tests)
}

func TestFunctionsWithReturnStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let earlyExit = function() { return 99; 100; };
			earlyExit();
			`,
			expected: 99,
		},
		{
			input: `
			let earlyExit = function() { return 99; return 100; };
			earlyExit();
			`,
			expected: 99,
		},
	}
	runVmTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let noReturn = function() { };
			noReturn();
			`,
			expected: Null,
		},
		{
			input: `
			let noReturn = function() { };
			let noReturnTwo = function() { noReturn(); };
			noReturn();
			noReturnTwo();
			`,
			expected: Null,
		},
	}
	runVmTests(t, tests)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let returnsOne = function() { 1; };
			let returnsOneReturner = function() { returnsOne; };
			returnsOneReturner()();
			`,
			expected: 1,
		},
		{
			input: `
			let returnsOneReturner = function() {
				let returnsOne = function() { 1; };
				returnsOne;
			};
			returnsOneReturner()();
			`,
			expected: 1,
		},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let one = function() { let one = 1; one };
			one();
			`,
			expected: 1,
		},
		{
			input: `
			let oneAndTwo = function() { let one = 1; let two = 2; one + two; };
			oneAndTwo();
			`,
			expected: 3,
		},
		{
			input: `
			let oneAndTwo = function() { let one = 1; let two = 2; one + two; };
			let threeAndFour = function() { let three = 3; let four = 4; three + four; };
			oneAndTwo() + threeAndFour();
			`,
			expected: 10,
		},
		{
			input: `
			let firstFoobar = function() { let foobar = 50; foobar; };
			let secondFoobar = function() { let foobar = 100; foobar; };
			firstFoobar() + secondFoobar();
			`,
			expected: 150,
		},
		{
			input: `
			let globalSeed = 50;
			let minusOne = function() {
				let num = 1;
				globalSeed - num;
			};
			let minusTwo = function() {
				let num = 2;
				globalSeed - num;
			};
			minusOne() + minusTwo();
			`,
			expected: 97,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let identity = function(a) { a; };
			identity(4);
			`,
			expected: 4,
		},
		{
			input: `
			let sum = function(a, b) { a + b; };
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let sum = function(a, b) {
				let c = a + b;
				c;
			};

			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let sum = function(a, b) {
				let c = a + b;
				c;
			};

			sum(1, 2) + sum(3, 4);`,
			expected: 10,
		},
		{
			input: `
			let sum = function(a, b) {
				let c = a + b;
				c;
			};

			let outer = function() {
				sum(1, 2) + sum(3, 4);
			};

			outer();
			`,
			expected: 10,
		},
		{
			input: `
			let globalNum = 10;

			let sum = function(a, b) {
				let c = a + b;
				c + globalNum;
			};

			let outer = function() {
				sum(1, 2) + sum(3, 4) + globalNum;
			};

			outer() + globalNum;
			`,
			expected: 50,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `function() { 1; }(1);`,
			expected: `wrong number of arguments: want=0, got=1`,
		},
		{
			input:    `function(a) { a; }();`,
			expected: `wrong number of arguments: want=1, got=0`,
		},
		{
			input:    `function(a, b) { a + b; }(1);`,
			expected: `wrong number of arguments: want=2, got=1`,
		},
		{
			input:    `function(a = 3) { 1; }(1, 2);`,
			expected: `wrong number of arguments: want=0-1, got=2`,
		},
		{
			input:    `function(a, b = 3) { 1; }();`,
			expected: `wrong number of arguments: want=1-2, got=0`,
		},
		{
			input:    `function(a, b, c = 3) { a + b; }(1);`,
			expected: `wrong number of arguments: want=2-3, got=1`,
		},
	}

	for _, tt := range tests {
		program := parse(t, tt.input)

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err == nil {
			t.Fatalf("expected VM error but resulted in none.")
		}

		if err.Error() != tt.expected {
			t.Fatalf("wrong VM error: want=%q, got=%q", tt.expected, err)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{
			`len(1)`,
			&object.Error{
				Message: "argument to `len` not supported, got INTEGER",
			},
		},
		{`len("one", "two")`,
			&object.Error{
				Message: "wrong number of arguments. got=2, want=1",
			},
		},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`println("hello", "world!")`, Null},
		{`first([1, 2, 3])`, 1},
		{`first([])`, Null},
		{`first(1)`,
			&object.Error{
				Message: "argument to `first` must be ARRAY, got INTEGER",
			},
		},
		{`last([1, 2, 3])`, 3},
		{`last([])`, Null},
		{`last(1)`,
			&object.Error{
				Message: "argument to `last` must be ARRAY, got INTEGER",
			},
		},
		{`rest([1, 2, 3])`, []int{2, 3}},
		{`rest([])`, Null},
		{`push([], 1)`, []int{1}},
		{`push(1, 1)`,
			&object.Error{
				Message: "argument to `push` must be ARRAY, got INTEGER",
			},
		},
	}

	runVmTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let newClosure = function(a) {
				function() { a; };
			};

			let closure = newClosure(99);
			closure();
			`,
			expected: 99,
		},
		{
			input: `
			let newAdder = function(a, b) {
				function(c) { a + b + c };
			};

			let adder = newAdder(1, 2);
			adder(8);
			`,
			expected: 11,
		},
		{
			input: `
			let newAdder = function(a, b) {
				let c = a + b;
				function(d) { c + d };
			};

			let adder = newAdder(1, 2);
			adder(8);
			`,
			expected: 11,
		},
		{
			input: `
			let newAdderOuter = function(a, b) {
				let c = a + b;

				function(d) {
					let e = d + c;
					function(f) { e + f; };
				};
			};

			let newAdderInner = newAdderOuter(1, 2);
			let adder = newAdderInner(3);
			adder(8);
			`,
			expected: 14,
		},
		{
			input: `
			let a = 1;

			let newAdderOuter = function(b) {
				function(c) {
					function(d) { a + b + c + d };
				};
			};

			let newAdderInner = newAdderOuter(2);
			let adder = newAdderInner(3);
			adder(8);
			`,
			expected: 14,
		},
		{
			input: `
			let newClosure = function(a, b) {
				let one = function() { a; };
				let two = function() { b; };

				function() { one() + two(); };
			};

			let closure = newClosure(9, 90);
			closure();
			`,
			expected: 99,
		},
	}

	runVmTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let countDown = function(x) {
				if (x == 0) {
					return 0;
				} else {
					countDown(x - 1);
				}
			};

			countDown(1);
			`,
			expected: 0,
		},
		{
			input: `
			let countDown = function(x) {
				if (x == 0) {
					return 0;
				} else {
					countDown(x - 1);
				}
			};

			let wrapper = function() {
				countDown(1);
			};

			wrapper();
			`,
			expected: 0,
		},
		{
			input: `
			let wrapper = function() {
				let countDown = function(x) {
					if (x == 0) {
						return 0;
					} else {
						countDown(x - 1);
					}
				};
				countDown(1);
			};

			wrapper();
			`,
			expected: 0,
		},
	}

	runVmTests(t, tests)
}

func TestRecursiveFibonacci(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let fibonacci = function(x) {
				if (x == 0) {
					return 0;
				} else {
					if (x == 1) {
						return 1;
					} else {
						fibonacci(x - 1) + fibonacci(x - 2);
					}
				}
			};

			fibonacci(15);
			`,
			expected: 610,
		},
	}
	runVmTests(t, tests)
}

func TestWhileLoopExpression(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let x = 1;
			let sum = 0;
			let up = 10;

			while (x < up) {
				sum += x;
				x++;
			}

			sum
			`,
			expected: 45,
		},
	}

	runVmTests(t, tests)
}

func TestAssignmentStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let a = 5; a += 1;", 6},
		{"let a = 5; a -= 1;", 4},
		{"let a = 6; a /= 2;", 3},
		{"let a = 6; a *= 2;", 12},
	}

	runVmTests(t, tests)
}

func TestPrefixAndPostfixStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let a = 5; a++;", 5},
		{"let a = 5; a++; a", 6},
		{"let a = 5; a--;", 5},
		{"let a = 5; a--; a", 4},
		{"let a = 5; ++a;", 6},
		{"let a = 5; --a;", 4},
	}

	runVmTests(t, tests)
}

func TestTailCalls(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			const factorial = function(n) {
			  if (n == 1) { return 1;}
			  n * factorial(n - 1);
			};
			factorial(5);`,
			expected: 120,
		},

		{
			input: `
			const factorial = function(n, a) {
			  if (n == 0) { return a;}
			  factorial(n - 1, a * n);
			};
			factorial(5, 1);`,
			expected: 120,
		},

		// without tail recursion optimization this will cause a stack overflow
		{
			input: `
			const iter = function(n, max) {
				if (n == max) {
					return n
				}
				return iter(n + 1, max)
			};
			iter(0, 9999)
			`,
			expected: 9999,
		},
	}

	runVmTests(t, tests)
}


/*
** Helpers
 */
func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(t, tt.input)

		opt, err := optimizer.New(program).Optimize()
		if err != nil {
			fmt.Printf("error while optimizing programm: %s", err)
			return
		}

		// fmt.Printf("\n\n Optimized:\n")
		// fmt.Printf("%s", opt.String())
	
		comp := compiler.New()
		err = comp.Compile(opt)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		// for i, constant := range comp.Bytecode().Constants {
		// 	fmt.Printf("CONSTANT %d %p (%T):\n", i, constant, constant)
		// 	switch constant := constant.(type) {
		// 	case *object.CompiledFunction:
		// 		fmt.Printf(" Instructions:\n%s", constant.Instructions)
		// 	case *object.Integer:
		// 		fmt.Printf(" Value: %d\n", constant.Value)
		// 	}
		// 	fmt.Printf("\n")
		// }

		// fmt.Printf("\n\n Instructions:\n")
		// fmt.Printf(comp.Bytecode().Instructions.String())

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
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

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}

	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}

	case *object.Null:
		if actual != Null {
			t.Errorf("object is not Null: %T (%+v)", actual, actual)
		}

	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}

	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object not Array: %T (%+v)", actual, actual)
			return
		}

		if len(array.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(array.Elements))
			return
		}

		for i, expectedElem := range expected {
			err := testIntegerObject(int64(expectedElem), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case map[object.HashKey]int64:
		hash, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d", len(expected), len(hash.Pairs))
			return
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				t.Errorf("no pair for given key in Pairs")
			}

			err := testIntegerObject(expectedValue, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case *object.Error:
		errObj, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", actual, actual)
			return
		}

		if errObj.Message != expected.Message {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, errObj.Message)
		}
	}
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer, got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%d, got=%d", expected, result.Value)
	}

	return nil
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("boolean was wrong value. want=%t got=%t", expected, result.Value)
	}

	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%q, want=%q", result.Value, expected)
	}

	return nil
}
