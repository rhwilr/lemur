package compiler

import (
	"encoding/binary"
	"fmt"

	"github.com/rhwilr/lemur/build"
	"github.com/rhwilr/lemur/code"
	"github.com/rhwilr/lemur/object"
)

const Signature = "rhwilr/lemur"

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

	header := writeHeader(uint16(len(b.Constants)), uint64(len(instructions)))

	out := append(header, constants...)
	out = append(out, instructions...)

	return out
}

func Read(bytecode []byte) (*Bytecode, error) {
	lenConstants, _, offset, err := readHeader(bytecode)
	if err != nil {
		return nil, err
	}

	constants, offset := readConstants(bytecode, offset, lenConstants)

	return &Bytecode{
		Constants:    constants,
		Instructions: bytecode[offset:],
	}, nil
}

/*
** Write Binary Functions
 */
func writeHeader(lenConstants uint16, lenInstructions uint64) []byte {
	out := make([]byte, 0)

	signature := []byte(Signature)
	version := []byte{build.BinaryVersion}

	constants := make([]byte, 2)
	binary.BigEndian.PutUint16(constants, lenConstants)
	instructions := make([]byte, 8)
	binary.BigEndian.PutUint64(instructions, lenInstructions)

	out = append(out, signature...)
	out = append(out, version...)
	out = append(out, constants...)
	out = append(out, instructions...)

	return out
}

func readHeader(bytecode []byte) (uint16, uint64, int, error) {
	offset := len(Signature)

	// Signature must be the magic value
	signature := string(bytecode[0:offset])

	if signature != Signature {
		return 0, 0, 0, fmt.Errorf("signature not found, expected '%s'", Signature)
	}

	version := bytecode[offset : offset+1]
	offset += 1
	if version[0] != build.BinaryVersion {
		return 0, 0, 0, fmt.Errorf("incompatible binary file version: vm=%02X bin=%02X", build.BinaryVersion, version[0])
	}

	constants := binary.BigEndian.Uint16(bytecode[offset : offset+2])
	offset += 2

	instructions := binary.BigEndian.Uint64(bytecode[offset : offset+8])
	offset += 8

	return constants, instructions, offset, nil
}

func writeConstants(consts []object.Object) []byte {
	out := newOutput()

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

			value := make([]byte, 16)
			binary.BigEndian.PutUint32(value[:], uint32(len(cnst.Instructions)))
			binary.BigEndian.PutUint32(value[4:], uint32(cnst.NumLocals))
			binary.BigEndian.PutUint32(value[8:], uint32(cnst.NumParameters))
			binary.BigEndian.PutUint32(value[12:], uint32(cnst.NumDefaults))

			value = append(value, cnst.Instructions...)
			out.write(byte(2), value)
		}
	}

	return out.Output
}

func readConstants(bytecode []byte, offset int, lenConstants uint16) ([]object.Object, int) {
	constants := make([]object.Object, 0)

	for index := 0; index < int(lenConstants); index++ {
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
			
			numDefaults := int(binary.BigEndian.Uint32(bytecode[offset : offset+4]))
			offset += 4

			instructions := bytecode[offset : offset+length]

			compiledFunctionObject := &object.CompiledFunction{
				NumLocals:     numLocals,
				NumParameters: numParameters,
				NumDefaults:   numDefaults,
				Instructions:  instructions,
			}

			constants = append(constants, compiledFunctionObject)

			offset += length
		}
	}

	return constants, offset
}

/*
** Helpers
 */
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
