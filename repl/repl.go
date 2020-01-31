package repl

import (
	"fmt"
	"io"
	"strings"

	"github.com/rhwilr/monkey/compiler"
	"github.com/rhwilr/monkey/lexer"
	"github.com/rhwilr/monkey/object"
	"github.com/rhwilr/monkey/parser"
	"github.com/rhwilr/monkey/vm"

	"github.com/carmark/pseudo-terminal-go/terminal"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	// Init monkey parser and vm
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)

	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	// Start repl
	term, err := terminal.NewWithStdInOut()
	if err != nil {
		panic(err)
	}
	defer term.ReleaseFromStdInOut() // defer this

	fmt.Println("Press Ctrl-D to break")
	term.SetPrompt(PROMPT)

	line, err := term.ReadLine()
	for {
		if err == io.EOF {
			fmt.Println()
			return
		}

		if (err != nil && strings.Contains(err.Error(), "control-c break")) || len(line) == 0 {
			line, err = term.ReadLine()
		} else {
			out := evaluateLine(line, symbolTable, constants, globals)

			term.Write([]byte(out + "\r\n"))
			line, err = term.ReadLine()
		}
	}
}

func evaluateLine(line string, symbolTable *compiler.SymbolTable, constants []object.Object, globals []object.Object) string {
	l := lexer.New(line)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(p.Errors())
		return ""
	}

	comp := compiler.NewWithState(symbolTable, constants)
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("Woops! Compilation failed:\n %s\n", err)
		return ""
	}

	code := comp.Bytecode()
	machine := vm.NewWithGlobalsStore(code, globals)

	err = machine.Run()
	if err != nil {
		fmt.Printf("Woops! Executing bytecode failed:\n %s\n", err)
		return ""
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped != nil {
		return lastPopped.Inspect()
	}

	return ""
}

func printParserErrors(errors []string) {
	fmt.Printf("Woops! We encountered a parse errors:\n")
	for _, msg := range errors {
		fmt.Printf("\t" + msg + "\n")
	}
}
