package compiler

type SymbolScope string

type Symbol struct {
	Name  string
	Index int
	Scope SymbolScope
}

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
)

type SymbolTable struct {
	Outer   *SymbolTable
	store   map[string]Symbol
	numDefs int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{store: make(map[string]Symbol)}
}

func (s *SymbolTable) Define(name string) Symbol {
	sym := Symbol{Name: name, Index: s.numDefs}
	if s.Outer == nil {
		sym.Scope = GlobalScope
	} else {
		sym.Scope = LocalScope
	}
	s.store[name] = sym
	s.numDefs++
	return sym
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	if sym, ok := s.store[name]; ok {
		return sym, true
	}
	if s.Outer != nil {
		return s.Outer.Resolve(name)
	}
	return Symbol{}, false
}

func (s *SymbolTable) NewEnclosed() *SymbolTable {
	st := NewSymbolTable()
	st.Outer = s
	return st
}
