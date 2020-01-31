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
	"path"

	"github.com/rhwilr/lemur/build"
	"github.com/rhwilr/lemur/compiler"
	"github.com/rhwilr/lemur/vm"
)

var (
	version bool
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [<filename>]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.BoolVar(&version, "v", false, "display version information")
}

func main() {
	flag.Parse()

	if version {
		fmt.Printf("%s %s\n", path.Base(os.Args[0]), build.FullVersion())
		os.Exit(0)
	}

	runVM()
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
		fmt.Printf("decode error: %s\n", err)
		return
	}

	machine := vm.New(code)

	err = machine.Run()
	if err != nil {
		fmt.Printf("vm error: %s\n", err)
		return
	}
}
