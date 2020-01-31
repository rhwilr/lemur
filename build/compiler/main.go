package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/rhwilr/monkey/build"
	"github.com/rhwilr/monkey/compiler"
	"github.com/rhwilr/monkey/lexer"
	"github.com/rhwilr/monkey/parser"
)

var (
	output  string
	version bool
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [<filename>]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.BoolVar(&version, "v", false, "display version information")
	flag.StringVar(&output, "o", "a.out", "name of the output file")
}

func main() {
	flag.Parse()

	if version {
		fmt.Printf("%s %s\n", path.Base(os.Args[0]), build.FullVersion())
		os.Exit(0)
	}

	runCompiler()
	os.Exit(0)
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
		log.Fatal(p.Errors())
	}

	c := compiler.New()
	err = c.Compile(program)
	if err != nil {
		log.Fatal(err)
	}

	code := c.Bytecode()

	bytecode := code.Write()
	f2, err := os.Create(output)
	if err != nil {
		fmt.Printf("vm error: %s", err)
		return
	}

	duration = time.Since(start)

	writer, _ := f2.Write(bytecode)
	fmt.Printf("wrote %d bytes in %s\n", writer, duration)
}
