package compiler

import (
	"github.com/rhwilr/monkey/code"
	"github.com/rhwilr/monkey/object"
)

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Constants:    c.constants,
		Instructions: c.currentInstructions(),
	}
}