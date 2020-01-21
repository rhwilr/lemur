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
- Allow string comparisons via `==`, `!=`, `>`, and `<`.


## Development

To setup the development environment for this repository you need `golang` installed. A `Makefile` is configured to run the most common tasks:

| Command      | Description                 |
| :----------- | :-------------------------- |
| `make test`  | Runs all tests.             |
| `make build` | Compiles the Monkey binary. |
