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

	// 上层符号表，nil 表示最外层，也就是 Global 层
	Outer *SymbolTable

	// 闭包函数所捕获的局部变量（即，在当前函数之外定义，且被当前函数所使用的局部变量）
	// 只有在编译用户自定义函数的过程中，逐个变量编译之后，才会产生完整的 FreeSymbols 列表。
	// 注意，
	// 来自上层的 Symbol 以 “原样” 的方式添加到 FreeSymbols（对于上 **1** 层来说，
	// 有些 Symbol 可能是 LocalScope），
	// 同时以 "FreeScope" 的方式添加到 store，把它当作当前符号表的一部分。
	FreeSymbols []Symbol
}

const (
	GlobalScope   SymbolScope = "GLOBAL"  // 全局变量
	BuiltinScope  SymbolScope = "BUILTIN" // 内置函数
	LocalScope    SymbolScope = "LOCAL"   // 局部变量
	FreeScope     SymbolScope = "FREE"    // 被闭包函数捕获的局部变量（该种变量的定义不在函数之内）
	FunctionScope SymbolScope = "FUNCTION"
)

func NewSymbolTable() *SymbolTable {
	store := make(map[string]Symbol)
	freeSymbols := []Symbol{}
	return &SymbolTable{store: store, FreeSymbols: freeSymbols}
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

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	// 把上一层的 Symbol 以 “原样” 的方式添加到 FreeSymbols
	s.FreeSymbols = append(s.FreeSymbols, original)

	// 再以 FreeScope 的形式把来自上层的 Symbol 添加到当前 store
	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1}
	symbol.Scope = FreeScope
	s.store[original.Name] = symbol
	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]

	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)

		// 符号未找到
		if !ok {
			return obj, ok
		}

		// 符号是全局（或内置函数）
		if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
			return obj, ok
		}

		// 符号既不是全局（或内置函数），也不是当前本地的，所以是 FreeScope
		// 只有当符号被访问时，才添加到 FreeSymbols 列表，也就是说，
		// 只有在编译用户自定义函数的过程中，逐个变量编译之后，才会产生完整的 FreeSymbols 列表。
		free := s.defineFree(obj)
		return free, true
	}

	return obj, ok
}

func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: FunctionScope}
	s.store[name] = symbol
	return symbol
}
