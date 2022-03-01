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
