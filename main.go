package main

// Package main implements the main process which executes a program if
// a filename is supplied as an argument or invokes the interpreter's
// REPL and waits for user input before lexing, parsing nad evaulating.

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/rhwilr/monkey/compiler"
	"github.com/rhwilr/monkey/lexer"
	"github.com/rhwilr/monkey/evaluator"
	"github.com/rhwilr/monkey/object"
	"github.com/rhwilr/monkey/parser"
	"github.com/rhwilr/monkey/repl"
	"github.com/rhwilr/monkey/vm"
)

var (
	engine      string
	interactive bool
	compile     bool
	version     bool
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [<filename>]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.BoolVar(&version, "v", false, "display version information")
	flag.BoolVar(&compile, "c", false, "compile input to bytecode")

	flag.BoolVar(&interactive, "i", false, "enable interactive mode")
	flag.StringVar(&engine, "e", "vm", "engine to use (eval or vm)")
}

func main() {
	flag.Parse()

	if version {
		fmt.Printf("%s %s\n", path.Base(os.Args[0]), FullVersion())
		os.Exit(0)
	}

	if compile {
		runCompiler()
		os.Exit(0)
	} 
	if interactive {
		runRepl()
		
		os.Exit(0)
	}
	
	runEvaluator()
	os.Exit(0)
}

func runRepl() {
	user, err := user.Current()
	if (err != nil) {
		panic(err)
	}

	fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}

func runEvaluator2() {
	args := flag.Args()

	f, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	input, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	var duration time.Duration
	var result object.Object

	l := lexer.New(string(input))
	p := parser.New(l)
	program := p.ParseProgram()

	if engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}

		machine := vm.New(comp.Bytecode())

		start := time.Now()

		err = machine.Run()
		if err != nil {
			fmt.Printf("vm error: %s", err)
			return
		}

		duration = time.Since(start)
		result = machine.LastPoppedStackElem()
	} else {
		env := object.NewEnvironment()
		start := time.Now()
		result = evaluator.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf(
		"engine=%s, result=%s, duration=%s\n",
		engine,
		result.Inspect(),
		duration)
}

func runEvaluator() {
	args := flag.Args()

	f, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var duration time.Duration

	input, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	code := compiler.Read(input)
	machine := vm.New(code)

	start := time.Now()

	err = machine.Run()
	if err != nil {
		fmt.Printf("vm error: %s", err)
		return
	}
	duration = time.Since(start)

	result := machine.LastPoppedStackElem()
	fmt.Printf("result=%s, duration=%s\n",
		result.Inspect(),
		duration)
}

func runCompiler() {
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("no source file given to compile")
	}
	
	f, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	l := lexer.New(string(b))
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		log.Fatal(p.Errors())
	}

	c := compiler.New()
	err = c.Compile(program)
	if err != nil {
		log.Fatal(err)
	}

	code := c.Bytecode()


	bytecode := code.Write()

	f2, err := os.Create("a.out")
	if err != nil {
		fmt.Printf("vm error: %s", err)
		return
	}

	n2, _ := f2.Write(bytecode)
	fmt.Printf("wrote %d bytes\n", n2)


	// run
	code = compiler.Read(bytecode)
	machine := vm.New(code)

	err = machine.Run()
	if err != nil {
		fmt.Printf("vm error: %s", err)
		return
	}

	result := machine.LastPoppedStackElem()
	fmt.Printf("result=%s\n",result.Inspect())
}
