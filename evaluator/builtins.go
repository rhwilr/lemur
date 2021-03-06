package evaluator

import (
	"github.com/rhwilr/lemur/object"
)

var builtins = map[string]*object.Builtin{
	"len":     object.GetBuiltinByName("len"),
	"read":    object.GetBuiltinByName("read"),
	"print":   object.GetBuiltinByName("print"),
	"println": object.GetBuiltinByName("println"),
	"first":   object.GetBuiltinByName("first"),
	"last":    object.GetBuiltinByName("last"),
	"rest":    object.GetBuiltinByName("rest"),
	"push":    object.GetBuiltinByName("push"),
	"env":     object.GetBuiltinByName("env"),
}
