package compiler

import (
	"encoding/binary"

	"github.com/rhwilr/monkey/code"
	"github.com/rhwilr/monkey/object"
)

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

type ConstantDefinition struct {
	Opcode        byte
	OperandWidths int
}

type Output struct {
	Output []byte
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Constants:    c.constants,
		Instructions: c.currentInstructions(),
	}
}

func (b *Bytecode) Write() []byte {
	constants := writeConstants(b.Constants)
	instructions := b.Instructions

	out := append(constants, instructions...)

	return out
}

func Read(bytecode []byte) *Bytecode {
	constants := make([]object.Object, 0)
	count := int(bytecode[0])

	offset := 1

	for index := 0; index < count; index++ {
		opcode := bytecode[offset]
		offset += 1

		switch opcode {
		case 0:
			length := 8
			value := int64(binary.BigEndian.Uint64(bytecode[offset : offset+length]))

			integerObject := &object.Integer{Value: value}
			constants = append(constants, integerObject)

			offset += length

		case 1:
			length := int(binary.BigEndian.Uint32(bytecode[offset : offset+4]))
			offset += 4
			value := string(bytecode[offset : offset+length])

			stringObject := &object.String{Value: value}
			constants = append(constants, stringObject)

			offset += length
	
		case 2:
			length := int(binary.BigEndian.Uint32(bytecode[offset : offset+4]))
			offset += 4

			numLocals := int(binary.BigEndian.Uint32(bytecode[offset : offset+4]))
			offset += 4

			numParameters := int(binary.BigEndian.Uint32(bytecode[offset : offset+4]))
			offset += 4

			instructions := bytecode[offset : offset+length]

			compiledFunctionObject := &object.CompiledFunction{
				NumLocals: numLocals,
				NumParameters: numParameters,
				Instructions: instructions,
			}

			constants = append(constants, compiledFunctionObject)

			offset += length
		}
	}

	return &Bytecode{
		Constants:    constants,
		Instructions: bytecode[offset:],
	}
}

func writeConstants(consts []object.Object) []byte {
	out := newOutput()

	var length int8 = int8(len(consts))
	out.write(byte(length), []byte{})

	for _, c := range consts {
		switch c.Type() {

		case object.INTEGER_OBJ:
			var cnst *object.Integer = c.(*object.Integer)

			value := make([]byte, 8)
			binary.BigEndian.PutUint64(value[:], uint64(cnst.Value))

			out.write(byte(0), value)

		case object.STRING_OBJ:
			var cnst *object.String = c.(*object.String)
			value := []byte(cnst.Value)

			length := make([]byte, 4)
			binary.BigEndian.PutUint32(length[:], uint32(len(value)))

			value = append(length, value...)
			out.write(byte(1), value)

		case object.COMPILED_FUNCTION_OBJ:
			var cnst *object.CompiledFunction = c.(*object.CompiledFunction)

			value := make([]byte, 12)
			binary.BigEndian.PutUint32(value[:], uint32(len(cnst.Instructions)))
			binary.BigEndian.PutUint32(value[4:], uint32(cnst.NumLocals))
			binary.BigEndian.PutUint32(value[8:], uint32(cnst.NumParameters))

			value = append(value, cnst.Instructions...)
			out.write(byte(2), value)
		}
	}

	return out.Output
}

// Output
func newOutput() *Output {
	return &Output{
		Output: make([]byte, 0),
	}
}

func (out *Output) write(op byte, operands []byte) {
	out.Output = append(out.Output, op)
	out.Output = append(out.Output, operands...)
}
