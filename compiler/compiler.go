package compiler

import (
	"fmt"
	"sort"

	"github.com/rhwilr/lemur/ast"
	"github.com/rhwilr/lemur/code"
	"github.com/rhwilr/lemur/object"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Compiler struct {
	constants []object.Object

	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	// Add builtins to the symbol table
	symbolTable := NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants

	return compiler
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

	case *ast.BlockStatement:
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

		// These instructions pop the last value, an additional OpPop is not
		// required.
		if !c.lastInstructionIs(code.OpSetGlobal) && !c.lastInstructionIs(code.OpSetLocal) {
			c.emit(code.OpPop)
		}

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit an `OpJumpNotTruthy` with a bogus value
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		// Remove the last pop, so the last value of the consequence is returned.
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		// Emit an `OpJump` with a bogus value
		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			// fix the opcode for the `OpJumpNotTruthy`
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.WhileLoopExpression:
		beforeConditionPos := len(c.currentInstructions())

		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit an `OpJumpNotTruthy` with a bogus value
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		// Remove the last pop, so the last value of the consequence is returned.
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		// Emit an `OpJump` with a bogus value
		c.emit(code.OpJump, beforeConditionPos)

		afterJumpPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterJumpPos)

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.PostfixExpression:
		err := c.Compile(node.Name)
		if err != nil {
			return err
		}

		symbol, ok := c.symbolTable.Resolve(node.Name.Value)
		if !ok {
			return fmt.Errorf("identifier not found: %s", node.Name.Value)
		}
		if symbol.Type == ConstantType {
			return fmt.Errorf("assignment to constant variable: %s", node.Name.Value)
		}

		c.loadSymbol(symbol)

		switch node.Operator {
		case "++":
			integer := &object.Integer{Value: 1}
			c.emit(code.OpConstant, c.addConstant(integer))
			c.emit(code.OpAdd)
		case "--":
			integer := &object.Integer{Value: 1}
			c.emit(code.OpConstant, c.addConstant(integer))
			c.emit(code.OpSub)
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpAssignGlobal, symbol.Index)
		} else {
			c.emit(code.OpAssignLocal, symbol.Index)
		}
		c.emit(code.OpPop)

	case *ast.InfixExpression:
		// The boolean operators AND and OR need special opcodes, since they should
		// only evaluate the second argument it the first did not short circuit.
		if node.Operator == "&&" || node.Operator == "||" {
			return c.compileLogicalInfixExpression(node)
		}

		// The < operator is not implemented in the VM, but we can use the >
		// operator by swapping the parameters.
		if node.Operator == "<" || node.Operator == "<=" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}

			switch node.Operator {
			case "<":
				c.emit(code.OpGreaterThan)
			case "<=":
				c.emit(code.OpGreaterOrEqual)
			}
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case ">=":
			c.emit(code.OpGreaterOrEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))

	case *ast.LetStatement:
		symbol, err := c.symbolTable.Define(node.Name.Value, VariableType)
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		err = c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}

	case *ast.ConstStatement:
		symbol, err := c.symbolTable.Define(node.Name.Value, ConstantType)
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		err = c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}

	// Assignment
	case *ast.AssignStatement:
		symbol, ok := c.symbolTable.Resolve(node.Name.Value)
		if !ok {
			return fmt.Errorf("identifier not found: %s", node.Name.Value)
		}
		if symbol.Type == ConstantType {
			return fmt.Errorf("assignment to constant variable: %s", node.Name.Value)
		}

		if node.Operator != "=" {
			c.loadSymbol(symbol)
		}

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+=", "++":
			c.emit(code.OpAdd)
		case "-=", "--":
			c.emit(code.OpSub)
		case "*=":
			c.emit(code.OpMul)
		case "/=":
			c.emit(code.OpDiv)
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpAssignGlobal, symbol.Index)
		} else {
			c.emit(code.OpAssignLocal, symbol.Index)
		}

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("identifier not found: %s", node.Value)
		}

		c.loadSymbol(symbol)

	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturn)

	case *ast.FunctionLiteral:
		c.enterScope()

		if node.Name != "" {
			c.symbolTable.DefineFunctionName(node.Name)
		}

		for _, p := range node.Parameters {
			symbol, err :=c.symbolTable.Define(p.Value, VariableType)

			if val, ok := node.Defaults[p.Value]; ok {
				if err != nil {
					return fmt.Errorf(err.Error())
				}

				currentInstructions := len(c.currentInstructions())
				err = c.Compile(val)
				if err != nil {
					return err
				}

				insertedInstructions := len(c.currentInstructions()) - currentInstructions
				for index := insertedInstructions; index < code.OptionalParameterInstructions; index++ {
					c.emit(code.OpNop)
				}

				c.emit(code.OpAssignLocal, symbol.Index)
			}
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}

		// If the function doesn't end with a return statement add one with a
		// `return null;` and also handle the edge-case of empty functions.
		if !c.lastInstructionIs(code.OpReturn) {
			// empty function body (LoadNull from BlockStatement)
			if !c.lastInstructionIs(code.OpNull) {
				c.emit(code.OpNull)
			}
			c.emit(code.OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()

		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}

		compiledFn := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
			NumDefaults: len(node.Defaults),
		}

		fnIndex := c.addConstant(compiledFn)
		c.emit(code.OpClosure, fnIndex, len(freeSymbols))

		if node.Define {
			symbol, err := c.symbolTable.Define(node.Name, VariableType)
			if err != nil {
				return fmt.Errorf(err.Error())
			}

			if symbol.Scope == GlobalScope {
				c.emit(code.OpSetGlobal, symbol.Index)
			} else {
				c.emit(code.OpSetLocal, symbol.Index)
			}
		}
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))
	}

	return nil
}

func (c *Compiler) compileLogicalInfixExpression(node *ast.InfixExpression) error {
	exp := &ast.IfExpression{}

	exp.Condition = node.Left

	// AND Expression
	if node.Operator == "&&" {
		consequenceBlock := &ast.BlockStatement{}
		consequenceBlock.Statements = []ast.Statement{
			&ast.ExpressionStatement{Expression: node.Right},
		}
		exp.Consequence = consequenceBlock

		alternativeBlock := &ast.BlockStatement{}
		alternativeBlock.Statements = []ast.Statement{
			&ast.ExpressionStatement{Expression: &ast.Boolean{Value: false}},
		}
		exp.Alternative = alternativeBlock
	}

	// OR Expression
	if node.Operator == "||" {
		consequenceBlock := &ast.BlockStatement{}
		consequenceBlock.Statements = []ast.Statement{
			&ast.ExpressionStatement{Expression: &ast.Boolean{Value: true}},
		}
		exp.Consequence = consequenceBlock

		alternativeBlock := &ast.BlockStatement{}
		alternativeBlock.Statements = []ast.Statement{
			&ast.ExpressionStatement{Expression: node.Right},
		}
		exp.Alternative = alternativeBlock
	}

	err := c.Compile(exp)
	if err != nil {
		return err
	}

	c.emit(code.OpCastToBool)

	return nil
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

/*
** Enter and leave scopes
 */
func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.Outer

	return instructions
}

/*
** Add, modify and remove instructions
 */
func (c *Compiler) addConstant(obj object.Object) int {
	switch obj := obj.(type) {
	case *object.Integer:
		for i, node := range c.constants {
			switch node := node.(type) {
			case *object.Integer:
				if obj.Value == node.Value {
					return i
				}
			}
		}

	case *object.String:
		for i, node := range c.constants {
			switch node := node.(type) {
			case *object.String:
				if obj.Value == node.Value {
					return i
				}
			}
		}
	}

	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturn))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturn
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	case FunctionScope:
		c.emit(code.OpCurrentClosure)
	}
}
