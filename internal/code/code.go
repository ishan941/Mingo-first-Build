package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

const (
	OpConstant Opcode = iota
	OpAdd
	OpSub
	OpMul
	OpDiv

	OpTrue
	OpFalse
	OpNull

	OpEqual
	OpNotEqual
	OpGreaterThan
	OpLessThan
	OpGreaterEqual
	OpLessEqual

	OpBang
	OpMinus

	OpPop

	OpJump
	OpJumpNotTruthy

	OpGetGlobal
	OpSetGlobal

	OpGetLocal
	OpSetLocal

	OpCall
	OpReturnValue
	OpReturn

	OpPrint
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant:      {Name: "OpConstant", OperandWidths: []int{2}},
	OpAdd:           {Name: "OpAdd"},
	OpSub:           {Name: "OpSub"},
	OpMul:           {Name: "OpMul"},
	OpDiv:           {Name: "OpDiv"},
	OpTrue:          {Name: "OpTrue"},
	OpFalse:         {Name: "OpFalse"},
	OpNull:          {Name: "OpNull"},
	OpEqual:         {Name: "OpEqual"},
	OpNotEqual:      {Name: "OpNotEqual"},
	OpGreaterThan:   {Name: "OpGreaterThan"},
	OpLessThan:      {Name: "OpLessThan"},
	OpGreaterEqual:  {Name: "OpGreaterEqual"},
	OpLessEqual:     {Name: "OpLessEqual"},
	OpBang:          {Name: "OpBang"},
	OpMinus:         {Name: "OpMinus"},
	OpPop:           {Name: "OpPop"},
	OpJump:          {Name: "OpJump", OperandWidths: []int{2}},
	OpJumpNotTruthy: {Name: "OpJumpNotTruthy", OperandWidths: []int{2}},
	OpGetGlobal:     {Name: "OpGetGlobal", OperandWidths: []int{2}},
	OpSetGlobal:     {Name: "OpSetGlobal", OperandWidths: []int{2}},
	OpGetLocal:      {Name: "OpGetLocal", OperandWidths: []int{1}},
	OpSetLocal:      {Name: "OpSetLocal", OperandWidths: []int{1}},
	OpCall:          {Name: "OpCall", OperandWidths: []int{1}},
	OpReturnValue:   {Name: "OpReturnValue"},
	OpReturn:        {Name: "OpReturn"},
	OpPrint:         {Name: "OpPrint"},
}

func Lookup(op Opcode) (*Definition, error) {
	def, ok := definitions[op]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	ins := make([]byte, instructionLen)
	ins[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(ins[offset:], uint16(o))
		case 1:
			ins[offset] = byte(o)
		}
		offset += width
	}

	return ins
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(binary.BigEndian.Uint16(ins[offset:]))
		case 1:
			operands[i] = int(ins[offset])
		}
		offset += width
	}
	return operands, offset
}

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		op := Opcode(ins[i])
		def, err := Lookup(op)
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	if len(operands) != len(def.OperandWidths) {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d", len(operands), len(def.OperandWidths))
	}

	switch len(operands) {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operand count for %s", def.Name)
}
