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

	"github.com/rhwilr/lemur/build"
	"github.com/rhwilr/lemur/compiler"
	"github.com/rhwilr/lemur/evaluator"
	"github.com/rhwilr/lemur/lexer"
	"github.com/rhwilr/lemur/optimizer"
	"github.com/rhwilr/lemur/object"
	"github.com/rhwilr/lemur/parser"
	"github.com/rhwilr/lemur/repl"
	"github.com/rhwilr/lemur/vm"
)

var (
	engine      string
	output      string
	interactive bool
	compile     bool
	execute     bool
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
	flag.BoolVar(&execute, "b", false, "execute a compiled file using the lemur-vm")
	flag.StringVar(&engine, "e", "vm", "engine to use (eval or vm), only supported with scripts")
	flag.StringVar(&output, "o", "a.out", "name of the output file")
}

func main() {
	flag.Parse()
	args := flag.Args()

	if version {
		fmt.Printf("%s %s\n", path.Base(os.Args[0]), build.FullVersion())
		os.Exit(0)
	}

	if compile {
		runCompiler()
		os.Exit(0)
	}

	if interactive || len(args) == 0 {
		runRepl()
		os.Exit(0)
	}

	if execute {
		runVM()
		os.Exit(0)
	}

	runEvaluator()
	os.Exit(0)
}

func runRepl() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! This is the Lemur programming language!\n", user.Username)
	repl.Start(os.Stdin, os.Stdout)
}

func runEvaluator() {
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
	if len(p.Errors()) != 0 {
		log.Fatalf("parse error: %s", p.Errors())
	}

	if engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			log.Fatalf("compiler error: %s", err)
		}

		machine := vm.New(comp.Bytecode())
		start := time.Now()

		err = machine.Run()
		if err != nil {
			log.Fatalf("vm error: %s", err)
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

func runCompiler() {
	args := flag.Args()
	var duration time.Duration
	start := time.Now()

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
		log.Fatalf("parse error: %s", p.Errors())
	}

	optimized, err := optimizer.New(program).Optimize()
	if err != nil {
		fmt.Printf("error while optimizing programm: %s", err)
		return
	}

	c := compiler.New()
	err = c.Compile(optimized)
	if err != nil {
		log.Fatalf("compiler error: %s", err)
	}

	code := c.Bytecode()

	bytecode := code.Write()
	f2, err := os.Create(output)
	if err != nil {
		log.Fatalf("compiler error: %s", err)
		return
	}

	duration = time.Since(start)

	writer, _ := f2.Write(bytecode)
	fmt.Printf("wrote %d bytes in %s\n", writer, duration)
}

func runVM() {
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

	code, err := compiler.Read(input)
	if err != nil {
		log.Fatalf("decode error: %s\n", err)
	}

	machine := vm.New(code)

	err = machine.Run()
	if err != nil {
		log.Fatalf("vm error: %s\n", err)
	}
}
