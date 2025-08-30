package vm

import (
	"errors"
	"fmt"

	"mingo/internal/code"
	"mingo/internal/object"
)

type VM struct {
	constants []object.Object

	globals []object.Object

	stack []object.Object
	sp    int // Always points to the next free slot on the stack

	ip           int
	instructions code.Instructions
}

const (
	StackSize   = 2048
	GlobalsSize = 65536
)

func New(instructions code.Instructions, constants []object.Object) *VM {
	return NewWithGlobals(instructions, constants, nil)
}

func NewWithGlobals(instructions code.Instructions, constants []object.Object, globals []object.Object) *VM {
	if globals == nil {
		globals = make([]object.Object, GlobalsSize)
	}
	return &VM{
		constants:    constants,
		globals:      globals,
		stack:        make([]object.Object, StackSize),
		sp:           0,
		ip:           0,
		instructions: instructions,
	}
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return errors.New("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	if vm.sp == 0 {
		return &object.Null{}
	}
	vm.sp--
	o := vm.stack[vm.sp]
	vm.stack[vm.sp] = nil
	return o
}

func (vm *VM) Run() error {
	for vm.ip < len(vm.instructions) {
		op := code.Opcode(vm.instructions[vm.ip])
		vm.ip++

		switch op {
		case code.OpConstant:
			idx := int(vm.instructions[vm.ip])<<8 | int(vm.instructions[vm.ip+1])
			vm.ip += 2
			if err := vm.push(vm.constants[idx]); err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			if err := vm.executeBinaryOperation(op); err != nil {
				return err
			}
		case code.OpTrue:
			if err := vm.push(&object.Boolean{Value: true}); err != nil {
				return err
			}
		case code.OpFalse:
			if err := vm.push(&object.Boolean{Value: false}); err != nil {
				return err
			}
		case code.OpNull:
			if err := vm.push(&object.Null{}); err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpGreaterEqual, code.OpLessThan, code.OpLessEqual:
			if err := vm.executeComparison(op); err != nil {
				return err
			}
		case code.OpBang:
			if err := vm.executeBang(); err != nil {
				return err
			}
		case code.OpMinus:
			right := vm.pop()
			i, ok := right.(*object.Integer)
			if !ok {
				return fmt.Errorf("unsupported negation operand %T", right)
			}
			if err := vm.push(&object.Integer{Value: -i.Value}); err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpJump:
			pos := int(vm.instructions[vm.ip])<<8 | int(vm.instructions[vm.ip+1])
			vm.ip = pos
		case code.OpJumpNotTruthy:
			pos := int(vm.instructions[vm.ip])<<8 | int(vm.instructions[vm.ip+1])
			vm.ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.ip = pos
			}
		case code.OpSetGlobal:
			idx := int(vm.instructions[vm.ip])<<8 | int(vm.instructions[vm.ip+1])
			vm.ip += 2
			vm.globals[idx] = vm.pop()
		case code.OpGetGlobal:
			idx := int(vm.instructions[vm.ip])<<8 | int(vm.instructions[vm.ip+1])
			vm.ip += 2
			if err := vm.push(vm.globals[idx]); err != nil {
				return err
			}
		case code.OpCall:
			argc := int(vm.instructions[vm.ip])
			vm.ip++
			fnObj := vm.stack[vm.sp-argc-1]
			fn, ok := fnObj.(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function: %T", fnObj)
			}
			// naive call: save current state on a small manual frame
			retIp := vm.ip
			retBase := vm.sp - argc - 1

			// new frame: execute function instructions
			savedIns := vm.instructions

			vm.instructions = fn.Instructions
			vm.ip = 0

			// locals are not fully supported; we just keep stack as is
			if err := vm.Run(); err != nil {
				return err
			}

			// function completed: top of stack has return value (or null)
			retVal := vm.pop()

			// restore
			vm.instructions = savedIns
			vm.ip = retIp

			// remove function and args, then push return value
			vm.sp = retBase
			if err := vm.push(retVal); err != nil {
				return err
			}
		case code.OpReturnValue:
			return nil
		case code.OpReturn:
			if err := vm.push(&object.Null{}); err != nil {
				return err
			}
			return nil
		case code.OpPrint:
			v := vm.pop()
			fmt.Println(v.Inspect())
		default:
			return fmt.Errorf("unsupported opcode: %d", op)
		}
	}
	return nil
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	li, lok := left.(*object.Integer)
	ri, rok := right.(*object.Integer)

	if !lok || !rok {
		return fmt.Errorf("unsupported types for binary op: %T %T", left, right)
	}

	var result int64
	switch op {
	case code.OpAdd:
		result = li.Value + ri.Value
	case code.OpSub:
		result = li.Value - ri.Value
	case code.OpMul:
		result = li.Value * ri.Value
	case code.OpDiv:
		if ri.Value == 0 {
			return errors.New("division by zero")
		}
		result = li.Value / ri.Value
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	li, lok := left.(*object.Integer)
	ri, rok := right.(*object.Integer)

	var result bool

	switch op {
	case code.OpEqual:
		if lok && rok {
			result = li.Value == ri.Value
		} else {
			result = left.Inspect() == right.Inspect()
		}
	case code.OpNotEqual:
		if lok && rok {
			result = li.Value != ri.Value
		} else {
			result = left.Inspect() != right.Inspect()
		}
	case code.OpGreaterThan:
		if !lok || !rok {
			return fmt.Errorf("> requires integers, got %T %T", left, right)
		}
		result = li.Value > ri.Value
	case code.OpGreaterEqual:
		if !lok || !rok {
			return fmt.Errorf(">= requires integers, got %T %T", left, right)
		}
		result = li.Value >= ri.Value
	case code.OpLessThan:
		if !lok || !rok {
			return fmt.Errorf("< requires integers, got %T %T", left, right)
		}
		result = li.Value < ri.Value
	case code.OpLessEqual:
		if !lok || !rok {
			return fmt.Errorf("<= requires integers, got %T %T", left, right)
		}
		result = li.Value <= ri.Value
	}

	return vm.push(&object.Boolean{Value: result})
}

func (vm *VM) executeBang() error {
	operand := vm.pop()
	switch v := operand.(type) {
	case *object.Boolean:
		return vm.push(&object.Boolean{Value: !v.Value})
	case *object.Null:
		return vm.push(&object.Boolean{Value: true})
	default:
		return vm.push(&object.Boolean{Value: false})
	}
}

func isTruthy(o object.Object) bool {
	switch v := o.(type) {
	case *object.Boolean:
		return v.Value
	case *object.Null:
		return false
	default:
		return true
	}
}
