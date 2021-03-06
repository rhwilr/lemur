package evaluator

import (
	"fmt"

	"github.com/rhwilr/lemur/ast"
	"github.com/rhwilr/lemur/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}

		return &object.ReturnValue{Value: val}
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

		// Expressions
	case *ast.BlockStatement:
		return evalBlockStatements(node.Statements, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		// Boolean operators
		if node.Operator == "&&" || node.Operator == "||" {
			return evalLogicalInfixExpression(node.Operator, node, env)
		}

		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)
	case *ast.PostfixExpression:
		return evalPostfixExpression(env, node.Operator, node)
	case *ast.IfExpression:
		return evalIfExpression(node, env)

	// LetStatements
	case *ast.LetStatement:
		if env.Exists(node.Name.Value, false) {
			return newError("identifier '%s' has already been declared", node.Name.Value)
		}

		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}

		env.DefineVariable(node.Name.Value, val)

	// ConstStatement
	case *ast.ConstStatement:
		if env.Exists(node.Name.Value, false) {
			return newError("identifier '%s' has already been declared", node.Name.Value)
		}

		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}

		env.DefineConstant(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &object.Array{Elements: elements}

	// Functions
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		defaults := node.Defaults

		function := &object.Function{Parameters: params, Env: env, Body: body, Defaults: defaults}

		// When the Define flag is set, the function should be registered in the env.
		if (node.Define) {
			if env.Exists(node.Name, false) {
				return newError("identifier '%s' has already been declared", node.Name)
			}
			
			env.DefineVariable(node.Name, function)
		}

		return function
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)

	case *ast.AssignStatement:
		return evalAssignStatement(node, env)
	case *ast.WhileLoopExpression:
		return evalWhileLoopExpression(node, env)
	}

	return nil
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalLogicalInfixExpression(operator string, node *ast.InfixExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	// AND operator
	if operator == "&&" {
		if !object.ObjectToNativeBoolean(left) {
			return nativeBoolToBooleanObject(false)
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return nativeBoolToBooleanObject(object.ObjectToNativeBoolean(right))
	}

	// OR operator
	if operator == "||" {
		if object.ObjectToNativeBoolean(left) {
			return nativeBoolToBooleanObject(true)
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return nativeBoolToBooleanObject(object.ObjectToNativeBoolean(right))
	}

	return newError("unknown boolean operator: %s", operator)
}

func evalPostfixExpression(env *object.Environment, operator string, node *ast.PostfixExpression) object.Object {
	val, ok := env.Get(node.Name.TokenLiteral())
	if !ok {
		return newError("%s is unknown", node.Name.TokenLiteral())
	}

	switch operator {
	case "++":
		switch arg := val.(type) {
		case *object.Integer:
			v := arg.Value
			_, err := env.Set(node.Name.TokenLiteral(), &object.Integer{Value: v + 1})
			if err != nil {
				return newError(err.Error())
			}
			return arg
		default:
			return newError("%s is not an int", node.Name.TokenLiteral())

		}
	case "--":
		switch arg := val.(type) {
		case *object.Integer:
			v := arg.Value
			_, err := env.Set(node.Name.TokenLiteral(), &object.Integer{Value: v - 1})
			if err != nil {
				return newError(err.Error())
			}
			return arg
		default:
			return newError("%s is not an int", node.Name.TokenLiteral())
		}
	default:
		return newError("unknown operator: %s", operator)
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "+", "+=":
		return &object.Integer{Value: leftVal + rightVal}
	case "-", "-=":
		return &object.Integer{Value: leftVal - rightVal}
	case "*", "*=":
		return &object.Integer{Value: leftVal * rightVal}
	case "/", "/=":
		return &object.Integer{Value: leftVal / rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "+":
		return &object.String{Value: leftVal + rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch isTruthy(right) {
	case false:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)

	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	case left.Type() == object.STRING_OBJ:
		return evalStringIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalStringIndexExpression(stringObj, index object.Object) object.Object {
	str := stringObj.(*object.String).Value
	idx := index.(*object.Integer).Value
	max := int64(len(str) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	chars := []rune(str)
	ret := chars[idx]

	return &object.String{Value: string(ret)}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		if result := fn.Fn(args...); result != nil {
			return result
		}
		return NULL
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func evalAssignStatement(a *ast.AssignStatement, env *object.Environment) (val object.Object) {
	evaluated := Eval(a.Value, env)
	if isError(evaluated) {
		return evaluated
	}

	current, ok := env.Get(a.Name.String())
	if !ok {
		return newError("assignment to undeclared variable '%s'", a.Name.String())
	}

	switch a.Operator {
	case "+=", "++":
		res := evalInfixExpression("+=", current, evaluated)
		if isError(res) {
			fmt.Printf("Error handling += %s\n", res.Inspect())
			return res
		}

		_, err := env.Set(a.Name.String(), res)
		if err != nil {
			return newError(err.Error())
		}
		return res

	case "-=", "--":
		res := evalInfixExpression("-=", current, evaluated)
		if isError(res) {
			fmt.Printf("Error handling -= %s\n", res.Inspect())
			return res
		}

		_, err := env.Set(a.Name.String(), res)
		if err != nil {
			return newError(err.Error())
		}
		return res

	case "*=":
		res := evalInfixExpression("*=", current, evaluated)
		if isError(res) {
			fmt.Printf("Error handling *= %s\n", res.Inspect())
			return res
		}

		_, err := env.Set(a.Name.String(), res)
		if err != nil {
			return newError(err.Error())
		}
		return res

	case "/=":
		res := evalInfixExpression("/=", current, evaluated)
		if isError(res) {
			fmt.Printf("Error handling /= %s\n", res.Inspect())
			return res
		}

		_, err := env.Set(a.Name.String(), res)
		if err != nil {
			return newError(err.Error())
		}
		return res

	case "=":
		// The assignment operator is not allowed to create new variables
		_, err := env.Set(a.Name.String(), evaluated)
		if err != nil {
			return newError(err.Error())
		}
	}
	return evaluated
}

func evalWhileLoopExpression(fle *ast.WhileLoopExpression, env *object.Environment) object.Object {
	rt := &object.Boolean{Value: true}

	for {
		condition := Eval(fle.Condition, env)

		if isError(condition) {
			return condition
		}

		if isTruthy(condition) {
			rt := Eval(fle.Consequence, env)

			if !isError(rt) && (rt.Type() == object.RETURN_VALUE_OBJ || rt.Type() == object.ERROR_OBJ) {
				return rt
			}
		} else {
			break
		}
	}

	return rt
}

/*
** Helpers
 */
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	// Set the defaults
	for key, val := range fn.Defaults {
		env.DefineVariable(key, Eval(val, env))
	}

	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.DefineVariable(param.Value, args[paramIdx])
		}
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func isTruthy(obj object.Object) bool {
	// Integer 0 is falsy
	if obj.Type() == object.INTEGER_OBJ && obj.(*object.Integer).Value == 0 {
		return false
	}

	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}

	return false
}
