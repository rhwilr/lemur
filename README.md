# Monkey

This repository contains an interpreter and compiler for the "Monkey"
programming language, as described in [Write an Interpreter in Go][1] and
[Writing A Compiler In Go][2] by Thorsten Ball.

[1]: https://interpreterbook.com/
[2]: https://compilerbook.com/


## Customizations

I made some changes in this implementation that differe from the implementation
in the book. Here are the changes I made.

- Added single-line & multi-line comments.
- Allow assignments without `let`.
- `let` can only be used to initialize a variable.
- More assignment operators such as `+=`, `-=`, `*=`, and `/=`.
- Added prefix and postfix operators (`++i`, `--i`, `i++`, `i--`).
- Allow accessing individual characters of a string via the index-operator.
- Allow string comparisons via `==`, `!=`, `<`, `>`, `<=`, and `>=`.
- Implemented `<=`, and `>=` comparisons for integers.
- Added support for boolean operators `&&` and `||`. This also adds support for
  more complex conditionals like `if (i <= 10 && containsNumber(string))...`.
- Implemented `while` loops.


## Installation

To use the most recent version, clone the source repository and run the make command to build the cli, compiler and vm:

```
git clone https://github.com/rhwilr/monkey.git
cd monkey
make
```

The latest binary can also be downloaded from
[Actions](https://github.com/rhwilr/monkey/actions) by clicking on the latest
build and downloading the artifacts.


## Usage

### Evaluator

To execute a script directly, use the `monkey` command line and pass the path to the script.

```
monkey examples/helo-world.mon
```

### Compiler

The `monke` executable can also be used to compile scripts with the `-c` flag.
However, there is also the `monkey-compiler` binary that does just that.

```
monkey-compiler examples/helo-world.mon
```

This will produce a binary file with the name `a.out` in the current folder. The
output file name can be changed with the `-o` parameter.

### VM

To execute the binary file, pass it to the `monkey-vm`:

```
monkey-vm a.out
```



## Development

To set up the development environment for this repository you need `golang` installed. A `Makefile` is configured to run the most common tasks:

| Command      | Description                 |
| :----------- | :-------------------------- |
| `make test`  | Runs all tests.             |
| `make build` | Compiles the Monkey binary. |
