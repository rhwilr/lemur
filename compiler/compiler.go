package compiler

import (
	"github.com/rhwilr/monkey/ast"
	"github.com/rhwilr/monkey/code"
	"github.com/rhwilr/monkey/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
		// Statements
		case *ast.Program:
			for _, s := range node.Statements {
				err := c.Compile(s)
				if err != nil {
					return err
				}
			}

		case *ast.ExpressionStatement:
			err := c.Compile(node.Expression)
			if err != nil {
				return err
			}

		case *ast.InfixExpression:
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}

			err = c.Compile(node.Right)
			if err != nil {
				return err
			}

		case *ast.IntegerLiteral:
			integer := &object.Integer{Value: node.Value}
			c.emit(code.OpConstant, c.addConstant(integer))
		}

		return nil
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) -1
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}