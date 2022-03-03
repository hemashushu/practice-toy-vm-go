package compiler

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		"b": Symbol{Name: "b", Scope: GlobalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	if a != expected["a"] {
		t.Errorf("expected 'a' %+v, actual %+v", expected["a"], a)
	}
	b := global.Define("b")
	if b != expected["b"] {
		t.Errorf("expected 'b' %+v, actual %+v", expected["b"], b)
	}
}

func TestResolveGlobal(t *testing.T) {
	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "b", Scope: GlobalScope, Index: 1},
	}

	global := NewSymbolTable()

	global.Define("a")
	global.Define("b")

	for _, sym := range expected {
		result, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, actual %+v",
				sym.Name, sym, result)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "b", Scope: GlobalScope, Index: 1},
		Symbol{Name: "c", Scope: LocalScope, Index: 0},
		Symbol{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, actual %+v",
				sym.Name, sym, result)
		}
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")
	secondLocal.Define("d") // 添加一个跟上一层同名的变量 // ++

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "c", Scope: LocalScope, Index: 0},
				Symbol{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "c", Scope: FreeScope, Index: 0}, // ++
				Symbol{Name: "e", Scope: LocalScope, Index: 0},
				Symbol{Name: "f", Scope: LocalScope, Index: 1},
				Symbol{Name: "d", Scope: LocalScope, Index: 2}, // ++
			},
		},
	}

	for _, test := range tests {
		for _, sym := range test.expectedSymbols {
			result, ok := test.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, actual %+v",
					sym.Name, sym, result)
			}
		}
	}
}

func TestDefine2(t *testing.T) {
	expected := map[string]Symbol{
		"a": Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		"b": Symbol{Name: "b", Scope: GlobalScope, Index: 1},
		"c": Symbol{Name: "c", Scope: LocalScope, Index: 0},
		"d": Symbol{Name: "d", Scope: LocalScope, Index: 1},
		"e": Symbol{Name: "e", Scope: LocalScope, Index: 0},
		"f": Symbol{Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	if a != expected["a"] {
		t.Errorf("expected a %+v, actual %+v", expected["a"], a)
	}

	b := global.Define("b")
	if b != expected["b"] {
		t.Errorf("expected b %+v, actual %+v", expected["b"], b)
	}

	firstLocal := NewEnclosedSymbolTable(global)

	c := firstLocal.Define("c")
	if c != expected["c"] {
		t.Errorf("expected c %+v, actual %+v", expected["c"], c)
	}

	d := firstLocal.Define("d")
	if d != expected["d"] {
		t.Errorf("expected d %+v, actual %+v", expected["d"], d)
	}

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	e := secondLocal.Define("e")
	if e != expected["e"] {
		t.Errorf("expected e %+v, actual %+v", expected["e"], e)
	}

	f := secondLocal.Define("f")
	if f != expected["f"] {
		t.Errorf("expected f %+v, actual %+v", expected["f"], f)
	}
}

func TestDefineResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	expected := []Symbol{
		Symbol{Name: "a", Scope: BuiltinScope, Index: 0},
		Symbol{Name: "c", Scope: BuiltinScope, Index: 1},
		Symbol{Name: "e", Scope: BuiltinScope, Index: 2},
		Symbol{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}

			if result != sym {
				t.Errorf("expected %s to resolve to %+v, actual %+v",
					sym.Name, sym, result)
			}
		}
	}
}

func TestResolveFree(t *testing.T) {

	// 下面构建的符号表所对应的程序如下
	//
	// let a = 1;
	// let b = 2;
	// let firstLocal = fn() {
	// 	let c = 3;
	// 	let d = 4;
	// 	a + b + c + d;
	// 	let secondLocal = fn() {
	// 		let e = 5;
	// 		let f = 6;
	// 		a + b + c + d + e + f;
	// 	};
	// };

	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "c", Scope: LocalScope, Index: 0},
				Symbol{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]Symbol{},
		},
		{
			secondLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "c", Scope: FreeScope, Index: 0},
				Symbol{Name: "d", Scope: FreeScope, Index: 1},
				Symbol{Name: "e", Scope: LocalScope, Index: 0},
				Symbol{Name: "f", Scope: LocalScope, Index: 1},
			},
			[]Symbol{
				Symbol{Name: "c", Scope: LocalScope, Index: 0},
				Symbol{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, test := range tests {
		for _, sym := range test.expectedSymbols {
			result, ok := test.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, actual %+v",
					sym.Name, sym, result)
			}
		}
		if len(test.table.FreeSymbols) != len(test.expectedFreeSymbols) {
			t.Errorf("wrong number of free symbols. expected %d, actual %d",
				len(test.expectedFreeSymbols),
				len(test.table.FreeSymbols),
			)
			continue
		}
		for i, sym := range test.expectedFreeSymbols {
			result := test.table.FreeSymbols[i]
			if result != sym {
				t.Errorf("wrong free symbol. expected %+v, actual %+v",
					sym, result)
			}
		}
	}
}

func TestResolveUnresolvableFree(t *testing.T) {

	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "c", Scope: FreeScope, Index: 0},
		Symbol{Name: "e", Scope: LocalScope, Index: 0},
		Symbol{Name: "f", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := secondLocal.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, actual %+v",
				sym.Name, sym, result)
		}
	}

	expectedUnresolvable := []string{
		"b",
		"d",
	}

	for _, name := range expectedUnresolvable {
		_, ok := secondLocal.Resolve(name)
		if ok {
			t.Errorf("name %s resolved, but was expected not to", name)
		}
	}
}

func TestDefineAndResolveFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")

	expected := Symbol{Name: "a", Scope: FunctionScope, Index: 0}

	result, ok := global.Resolve(expected.Name)
	if !ok {
		t.Fatalf("function name %s not resolvable", expected.Name)
	}
	if result != expected {
		t.Errorf("expected %s to resolve to %+v, actual %+v",
			expected.Name, expected, result)
	}
}

func TestShadowingFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")
	global.Define("a")

	expected := Symbol{Name: "a", Scope: GlobalScope, Index: 0}

	result, ok := global.Resolve(expected.Name)
	if !ok {
		t.Fatalf("function name %s not resolvable", expected.Name)
	}
	if result != expected {
		t.Errorf("expected %s to resolve to %+v, actual %+v",
			expected.Name, expected, result)
	}
}
