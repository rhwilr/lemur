package object

import "fmt"

type Environment struct {
	variables map[string]Object
	constants map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Object), 
		constants: make(map[string]Object), 
		outer: nil,
	}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer

	return env
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.constants[name]

	//constant was found
	if ok {
		return obj, ok
	}

	obj, ok = e.variables[name]
	// variable was found
	if ok {
		return obj, ok
	}

	// no outer environment
	if (e.outer == nil) {
		return nil, false
	}

	return e.outer.Get(name)
}

func (e *Environment) Exists(name string) (bool) {
	return e.constantExists(name) || e.variableExists(name)
}

func (e * Environment) Set(name string, val Object) (Object, error) {
	if e.constantExists(name) {
		return val, fmt.Errorf("assignment to constant variable '%s'", name)
	}

	if !e.variableExists(name) {
		return val, fmt.Errorf("assignment to undeclared variable '%s'", name)
	}

	e.variables[name] = val

	return val, nil
}

func (e * Environment) DefineConstant(name string, val Object) Object {
	e.constants[name] = val
	return val
}

func (e * Environment) DefineVariable(name string, val Object) Object {
	e.variables[name] = val
	return val
}

func (e *Environment) constantExists(name string) (bool) {
	_, ok := e.constants[name]
	return ok
}

func (e *Environment) variableExists(name string) (bool) {
	_, ok := e.variables[name]
	return ok
}
