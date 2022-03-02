package compiler

type SymbolScope string

// 符号/标识符
type Symbol struct {
	Name  string      // 符号的名称
	Scope SymbolScope // 符号的范围
	Index int         // 符号的索引
}

type SymbolTable struct {
	store          map[string]Symbol // 存储符号的记录（用 map 实现）
	numDefinitions int               // 符号的数量

	Outer *SymbolTable // 上层符号表，nil 表示最外层，也就是 Global 层
}

const (
	GlobalScope  SymbolScope = "GLOBAL"
	BuiltinScope SymbolScope = "BUILTIN"
	LocalScope   SymbolScope = "LOCAL"
)

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Index: s.numDefinitions, // 使用当前记录数量作为符号的索引值
		// Scope: GlobalScope,
	}

	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]

	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		// return obj, ok
	}

	return obj, ok
}
