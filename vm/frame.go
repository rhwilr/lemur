package vm

import (
	"github.com/rhwilr/lemur/code"
	"github.com/rhwilr/lemur/object"
)

type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer int
}

func NewFrame(cl *object.Closure, basePointer int, ip int) *Frame {
	f := &Frame{
		cl:          cl,
		ip:          ip,
		basePointer: basePointer,
	}

	return f
}

// NextOp ...
func (f *Frame) NextOp() code.Opcode {
	return code.Opcode(f.Instructions()[f.ip+1])
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
