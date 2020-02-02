package optimizer

import (
	"strconv"

	"github.com/rhwilr/lemur/ast"
	"github.com/rhwilr/lemur/token"
)

type Optimizer struct {
	program *ast.Program
	optimized *ast.Program
}

func New(program *ast.Program) *Optimizer {
	return &Optimizer{
		program: program,
		optimized: &ast.Program{},
	}
}

func (o *Optimizer) Optimize() (*ast.Program, error) {
	for _, statement := range o.program.Statements {
		o.optimizeStatement(statement)
	}

	return o.optimized, nil
}

func (o *Optimizer) optimizeStatement(statement ast.Statement) {
	switch statement := statement.(type) {
	case *ast.LetStatement:
		value := optimizeLetStatement(statement.Value)

		if (value != nil) {
			statement.Value = value
		}

		o.optimized.Statements = append(o.optimized.Statements, statement)

	case *ast.ExpressionStatement:
		value, _ := evaluateExpression(statement.Expression)

		statement.Expression = value
		o.optimized.Statements = append(o.optimized.Statements, statement)
	default:
		o.optimized.Statements = append(o.optimized.Statements, statement)
	}
}

func optimizeLetStatement(node ast.Expression) ast.Expression {
	switch node := node.(type) {
	case *ast.InfixExpression:
		return optimizeInfixExpression(node)
	}

	return node
}

func optimizeInfixExpression(node *ast.InfixExpression) ast.Expression {
	left, ok := evaluateExpression(node.Left)
	if !ok {
		return node
	}

	right, ok := evaluateExpression(node.Right)
	if !ok {
		return node
	}

	// Integers
	_, okL := left.(*ast.IntegerLiteral)
	_, okR := right.(*ast.IntegerLiteral)
	if okL && okR {
		return optimizeIntegerInfixExpression(node.Operator, left, right)
	}

	// Strings
	_, okL = left.(*ast.StringLiteral)
	_, okR = right.(*ast.StringLiteral)
	if okL && okR {
		return optimizeStringInfixExpression(node.Operator, left, right)
	}

	return node
}


func optimizeIntegerInfixExpression(operator string, left, right ast.Expression) ast.Expression {
	leftVal := left.(*ast.IntegerLiteral).Value
	rightVal := right.(*ast.IntegerLiteral).Value

	switch operator {
	case "==":
		return nativeBoolToBooleanAst(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanAst(leftVal != rightVal)
	case "<":
		return nativeBoolToBooleanAst(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanAst(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanAst(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanAst(leftVal >= rightVal)
	case "+", "+=":
		return nativeIntegerToIntegerAst(leftVal + rightVal)
	case "-", "-=":
		return nativeIntegerToIntegerAst(leftVal - rightVal)
	case "*", "*=":
		return nativeIntegerToIntegerAst(leftVal * rightVal)
	case "/", "/=":
		return nativeIntegerToIntegerAst(leftVal / rightVal)
	default:
		return nil
	}
}

func optimizeStringInfixExpression(operator string, left, right ast.Expression) ast.Expression {
	leftVal := left.(*ast.StringLiteral).Value
	rightVal := right.(*ast.StringLiteral).Value

	switch operator {
	case "==":
		return nativeBoolToBooleanAst(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanAst(leftVal != rightVal)
	case "<":
		return nativeBoolToBooleanAst(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanAst(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanAst(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanAst(leftVal >= rightVal)
	case "+":
		return nativeStringToStringAst(leftVal + rightVal)
	default:
		return nil
	}
}

func optimizeWhileLoopExpression(node *ast.WhileLoopExpression) ast.Expression {
	condition, _ := evaluateExpression(node.Condition)

	node.Condition = condition
	
	return node
}

func optimizeCallExpression(node *ast.CallExpression) ast.Expression {
	list := []ast.Expression{}

	for _, argument := range node.Arguments {
		optimized, _ := evaluateExpression(argument)
		list = append(list, optimized)
	}

	node.Arguments = list

	return node
}

func evaluateExpression(node ast.Expression) (ast.Expression, bool) {
	switch node := node.(type) {
	case *ast.IntegerLiteral:
		return node, true
	case *ast.Boolean:
		return node, true
	case *ast.StringLiteral:
		return node, true
	case *ast.InfixExpression:
		return optimizeInfixExpression(node), true
	case *ast.WhileLoopExpression:
		return optimizeWhileLoopExpression(node), true
	case *ast.CallExpression:
		return optimizeCallExpression(node), true
	}

	return node, false 
}

/*
** Helper
*/
func nativeBoolToBooleanAst(input bool) *ast.Boolean {
	if input {
		return &ast.Boolean{
			Token: token.Token{
				Type: token.TRUE,
				Literal: "true",
			},
			Value: input,
		}
	}

	return &ast.Boolean{
		Token: token.Token{
			Type: token.FALSE,
			Literal: "false",
		},
		Value: input,
	}
}

func nativeIntegerToIntegerAst(value int64) *ast.IntegerLiteral {
	return &ast.IntegerLiteral{
		Token: token.Token{
			Type:token.INT,
			Literal: strconv.FormatInt(value, 10),
		},
		Value: value,
	}
}

func nativeStringToStringAst(value string) *ast.StringLiteral {
	return &ast.StringLiteral{
		Token: token.Token{
			Type:token.STRING,
			Literal: value,
		},
		Value: value,
	}
}