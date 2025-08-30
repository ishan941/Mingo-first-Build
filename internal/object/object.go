package object

import (
	"fmt"
	"strings"
)

type Type string

const (
	INTEGER_OBJ           Type = "INTEGER"
	BOOLEAN_OBJ           Type = "BOOLEAN"
	NULL_OBJ              Type = "NULL"
	COMPILED_FUNCTION_OBJ Type = "COMPILED_FUNCTION"
)

type Object interface {
	Type() Type
	Inspect() string
}

type Integer struct{ Value int64 }

func (i *Integer) Type() Type      { return INTEGER_OBJ }
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

type Boolean struct{ Value bool }

func (b *Boolean) Type() Type { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}

type Null struct{}

func (n *Null) Type() Type      { return NULL_OBJ }
func (n *Null) Inspect() string { return "null" }

type CompiledFunction struct {
	Instructions  []byte
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() Type { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	var b strings.Builder
	b.WriteString("compiled fn[")
	b.WriteString(fmt.Sprintf("params=%d locals=%d", cf.NumParameters, cf.NumLocals))
	b.WriteString("]")
	return b.String()
}
