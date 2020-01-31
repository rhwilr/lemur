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
- Added support for logical operators `&&` and `||`. This also adds support for
  more complex conditionals like `if (i <= 10 && containsNumber(string))...`.
- Implemented `while` loops.


## Installation

To use the most recent version, clone the source repository and run the make command to build the cli, compiler and vm:

```sh
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

```sh
monkey examples/helo-world.mon
```

### Compiler

The `monke` executable can also be used to compile scripts with the `-c` flag.
However, there is also the `monkey-compiler` binary that does just that.

```sh
monkey-compiler examples/helo-world.mon
```

This will produce a binary file with the name `a.out` in the current folder. The
output file name can be changed with the `-o` parameter.

### VM

To execute the binary file, pass it to the `monkey-vm`:

```sh
monkey-vm a.out
```


## Syntax

**Note:** You can find some example programms in the [examples](examples/)
folder.


## Data Types

Monkey has support for the following data types:
- Integer
- Boolen
- String
- Array
- Hash
- Null
- Functions

Yes, functions are a first class citizen and can be passed as arguments or
returned from other functions.


### Definitions

We have support for constants and variables:

```js
const a = 3;
let number = 7;
```

Variables can be updated using asignments. Constants can not be updated.

```js
number = 8;

// prefix and postfix operators
// -- is also supported
number++;       // returns the current value, then increments number by 1
++number;       // increments number by 1, then returns the new value

number += 5;    // Adds 5 to the number
```


### Arithmetic operations

Monkey supports all the basic arithmetic operations of Integer types.

```js
let a = 4;
let b = 2;

puts( a + b );  // Outputs: 6
puts( a - b );  // Outputs: 2
puts( a * b );  // Outputs: 8
puts( a / b );  // Outputs: 2
```


### Builtin functions

These core primitives are part of the monkey language:

- `len`
- `first`
- `last`
- `rest`
- `push`
- `puts`


### Container Datatypes

### Conditionals

### While-loops

### Comments

### Functions



## Development

To set up the development environment for this repository you need `golang` installed. A `Makefile` is configured to run the most common tasks:

| Command      | Description                 |
| :----------- | :-------------------------- |
| `make test`  | Runs all tests.             |
| `make build` | Compiles the Monkey binary. |
