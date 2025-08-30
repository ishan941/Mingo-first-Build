package compiler

import (
	"fmt"

	"mingo/internal/ast"
	"mingo/internal/code"
	"mingo/internal/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object

	symTable *SymbolTable
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
		symTable:     NewSymbolTable(),
	}
}

// NewWithState creates a compiler that reuses an existing symbol table and constants.
func NewWithState(sym *SymbolTable, consts []object.Object) *Compiler {
	if sym == nil {
		sym = NewSymbolTable()
	}
	if consts == nil {
		consts = []object.Object{}
	}
	return &Compiler{
		instructions: code.Instructions{},
		constants:    consts,
		symTable:     sym,
	}
}

func (c *Compiler) Instructions() code.Instructions { return c.instructions }
func (c *Compiler) Constants() []object.Object      { return c.constants }

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return pos
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) Compile(node ast.Node) error {
	switch n := node.(type) {
	case *ast.Program:
		for _, s := range n.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		if err := c.Compile(n.Expression); err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.IntegerLiteral:
		i := &object.Integer{Value: n.Value}
		constIdx := c.addConstant(i)
		c.emit(code.OpConstant, constIdx)
	case *ast.Boolean:
		if n.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.PrefixExpression:
		if err := c.Compile(n.Right); err != nil {
			return err
		}
		switch n.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %q", n.Operator)
		}
	case *ast.InfixExpression:
		// Order matters for < and <=; compile as > or >= by swapping
		if n.Operator == "<" || n.Operator == "<=" {
			if err := c.Compile(n.Right); err != nil {
				return err
			}
			if err := c.Compile(n.Left); err != nil {
				return err
			}
			if n.Operator == "<" {
				c.emit(code.OpGreaterThan)
			} else {
				c.emit(code.OpGreaterEqual)
			}
			return nil
		}
		if err := c.Compile(n.Left); err != nil {
			return err
		}
		if err := c.Compile(n.Right); err != nil {
			return err
		}
		switch n.Operator {
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
		case ">=":
			c.emit(code.OpGreaterEqual)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %q", n.Operator)
		}
	case *ast.LetStatement:
		if err := c.Compile(n.Value); err != nil {
			return err
		}
		sym := c.symTable.Define(n.Name.Value)
		switch sym.Scope {
		case GlobalScope:
			c.emit(code.OpSetGlobal, sym.Index)
		case LocalScope:
			c.emit(code.OpSetLocal, sym.Index)
		}
	case *ast.Identifier:
		sym, ok := c.symTable.Resolve(n.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", n.Value)
		}
		switch sym.Scope {
		case GlobalScope:
			c.emit(code.OpGetGlobal, sym.Index)
		case LocalScope:
			c.emit(code.OpGetLocal, sym.Index)
		}
	case *ast.AssignmentStatement:
		// compile RHS then assign to existing symbol
		if err := c.Compile(n.Value); err != nil {
			return err
		}
		sym, ok := c.symTable.Resolve(n.Name.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", n.Name.Value)
		}
		switch sym.Scope {
		case GlobalScope:
			c.emit(code.OpSetGlobal, sym.Index)
		case LocalScope:
			c.emit(code.OpSetLocal, sym.Index)
		}
	case *ast.BlockStatement:
		for _, s := range n.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
	case *ast.IfExpression:
		if err := c.Compile(n.Condition); err != nil {
			return err
		}
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)
		if err := c.Compile(n.Consequence); err != nil {
			return err
		}
		jumpPos := c.emit(code.OpJump, 9999)
		// patch jump-not-truthy to after consequence
		afterConsequence := len(c.instructions)
		c.replaceOperand(jumpNotTruthyPos, afterConsequence)
		if n.Alternative != nil {
			if err := c.Compile(n.Alternative); err != nil {
				return err
			}
		}
		afterAlternative := len(c.instructions)
		c.replaceOperand(jumpPos, afterAlternative)
	case *ast.WhileStatement:
		loopStart := len(c.instructions)
		if err := c.Compile(n.Condition); err != nil {
			return err
		}
		exitJumpPos := c.emit(code.OpJumpNotTruthy, 9999)
		if err := c.Compile(n.Body); err != nil {
			return err
		}
		c.emit(code.OpJump, loopStart)
		afterLoop := len(c.instructions)
		c.replaceOperand(exitJumpPos, afterLoop)
	case *ast.FunctionLiteral:
		// Enter function scope
		outer := c.symTable
		c.symTable = outer.NewEnclosed()

		for _, p := range n.Parameters {
			c.symTable.Define(p.Value)
		}
		// Compile body
		if err := c.Compile(n.Body); err != nil {
			return err
		}
		c.emit(code.OpReturn)

		fn := &object.CompiledFunction{Instructions: c.instructions, NumLocals: c.symTable.numDefs, NumParameters: len(n.Parameters)}

		// restore
		c.instructions = code.Instructions{}
		c.symTable = outer

		idx := c.addConstant(fn)
		c.emit(code.OpConstant, idx)
	case *ast.FunctionStatement:
		// Compile like a function literal, then assign to name
		lit := &ast.FunctionLiteral{Token: n.Token, Parameters: n.Parameters, Body: n.Body}
		if err := c.Compile(lit); err != nil {
			return err
		}
		sym := c.symTable.Define(n.Name.Value)
		c.emit(code.OpSetGlobal, sym.Index)
	case *ast.CallExpression:
		if err := c.Compile(n.Function); err != nil {
			return err
		}
		for _, a := range n.Arguments {
			if err := c.Compile(a); err != nil {
				return err
			}
		}
		c.emit(code.OpCall, len(n.Arguments))
	case *ast.ReturnStatement:
		if n.ReturnValue != nil {
			if err := c.Compile(n.ReturnValue); err != nil {
				return err
			}
			c.emit(code.OpReturnValue)
		} else {
			c.emit(code.OpReturn)
		}
	case *ast.PrintStatement:
		if err := c.Compile(n.Value); err != nil {
			return err
		}
		c.emit(code.OpPrint)
	default:
		return fmt.Errorf("unhandled node type %T", n)
	}

	return nil
}

func (c *Compiler) replaceOperand(pos int, operand int) {
	op := code.Opcode(c.instructions[pos])
	operands := []int{operand}
	newIns := code.Make(op, operands...)
	c.replaceInstruction(pos, newIns)
}

func (c *Compiler) replaceInstruction(pos int, newIns []byte) {
	for i := 0; i < len(newIns); i++ {
		c.instructions[pos+i] = newIns[i]
	}
}
