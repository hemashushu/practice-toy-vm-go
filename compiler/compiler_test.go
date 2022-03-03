package compiler

import (
	"fmt"
	"testing"
	"toyvm/ast"
	"toyvm/code"
	"toyvm/lexer"
	"toyvm/object"
	"toyvm/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, test := range tests {
		program := parse(test.input)
		compiler := New()

		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		// 检查字节码的指令部分（.text），一个 byte 数组
		err = testInstructions(test.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		// 检查字节码的常数部分（.data），一个 byte 数组
		err = testConstants(t, test.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

// 检查字节码的指令部分（.text），一个 byte 数组
func testInstructions(expected []code.Instructions, actual code.Instructions) error {

	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("instructions length expected\n%q, actual\n%q",
			concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("[byte %d] instruction expected\n%q, actual\n%q",
				i, concatted, actual)
		}
	}
	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}
	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}

// 检查字节码的常数部分（.data），一个 byte 数组
func testConstants(t *testing.T, expected []interface{}, actual []object.Object) error {

	if len(expected) != len(actual) {
		return fmt.Errorf("number of constants expected %d, actual %d",
			len(actual), len(expected))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("[%d] testIntegerObject failed: %s",
					i, err)
			}

		case string: // 添加对 String 的支持
			err := testStringObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf("[%d] testStringObject failed: %s", i, err)
			}

		case []code.Instructions:
			fn, ok := actual[i].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("[%d] not a function: %T",
					i, actual[i])
			}

			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("[%d] testInstructions failed: %s",
					i, err)
			}
		}
	}

	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer, actual %T, %+v",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value, expected %d, actual %d",
			expected, result.Value)
	}

	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String, actual %T, %+v",
			actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value, expected %q, actual %q",
			expected, result.Value)
	}
	return nil
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 * 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "2 / 1",
			expectedConstants: []interface{}{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []interface{}{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			if (true) { 10 }; 3333;
			`,
			expectedConstants: []interface{}{10, 3333},
			expectedInstructions: []code.Instructions{
				/* 0000 */ code.Make(code.OpTrue), // 1 bytes

				// 在支持 Null 之前的 if 语句
				//
				// /* 0001 */ code.Make(code.OpJumpNotTruthy, 7), // 3 bytes
				// /* 0004 */ code.Make(code.OpConstant, 0), // 3 bytes
				// /* 0007 */ code.Make(code.OpPop), // 1 bytes
				// /* 0008 */ code.Make(code.OpConstant, 1), // 3 bytes
				// /* 0011 */ code.Make(code.OpPop), // 1 bytes

				// 支持 Null 之后的 if 语句
				/* 0001 */ code.Make(code.OpJumpNotTruthy, 10), // 3 bytes
				/* 0004 */ code.Make(code.OpConstant, 0), // 3 bytes
				/* 0007 */ code.Make(code.OpJump, 11), // 3 bytes
				/* 0010 */ code.Make(code.OpNull), // 1 bytes
				/* 0011 */ code.Make(code.OpPop), // 1 bytes ;; 清理 if 语句的值
				/* 0012 */ code.Make(code.OpConstant, 1), // 3 bytes
				/* 0015 */ code.Make(code.OpPop), // 1 bytes
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals2(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			if (true) { 10 } else { 20 }; 3333;
			`,
			expectedConstants: []interface{}{10, 20, 3333},
			expectedInstructions: []code.Instructions{
				/* 0000 */ code.Make(code.OpTrue), // 1 bytes
				/* 0001 */ code.Make(code.OpJumpNotTruthy, 10), // 3 bytes
				/* 0004 */ code.Make(code.OpConstant, 0), // 3 bytes, "10" 语句
				/* 0007 */ code.Make(code.OpJump, 13), // 3 bytes
				/* 0010 */ code.Make(code.OpConstant, 1), // 3 bytes, "20" 语句
				/* 0013 */ code.Make(code.OpPop), // 1 bytes, "if..." 语句的结束
				/* 0014 */ code.Make(code.OpConstant, 2), // 3 bytes
				/* 0017 */ code.Make(code.OpPop), // 1 bytes, "3333;" 语句的结束
			},
		},
	}

	runCompilerTests(t, tests)
}

// 测试当 if 语句的条件不成立，但又缺少 alternative 语句块的情况
func TestConditionals3(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			if (true) { 10 }; 3333;
			`,
			expectedConstants: []interface{}{10, 3333},
			expectedInstructions: []code.Instructions{
				/* 0000 */ code.Make(code.OpTrue), // 1 bytes
				/* 0001 */ code.Make(code.OpJumpNotTruthy, 10), // 3 bytes
				/* 0004 */ code.Make(code.OpConstant, 0), // 3 bytes, ;; "10"
				/* 0007 */ code.Make(code.OpJump, 11), // 3 bytes
				/* 0010 */ code.Make(code.OpNull), // 1 bytes ;; 因为缺少 alternative 语句块而补上的指令
				/* 0011 */ code.Make(code.OpPop), // 1 bytes
				/* 0012 */ code.Make(code.OpConstant, 1), // 3 bytes, ;; "3333"
				/* 0015 */ code.Make(code.OpPop), // 1 bytes
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			let one = 1;
			let two = 2;
			`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `
			let one = 1;
			one;
			`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let one = 1;
			let two = one;
			two;
			`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"monkey"`,
			expectedConstants: []interface{}{"monkey"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"mon" + "key"`,
			expectedConstants: []interface{}{"mon", "key"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1 + 2, 3 - 4, 5 * 6]",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpHash, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2, 3: 4, 5: 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpHash, 6),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2 + 3, 4: 5 * 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpHash, 4),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1, 2, 3][1 + 1]",
			expectedConstants: []interface{}{1, 2, 3, 1, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2}[2 - 1]",
			expectedConstants: []interface{}{1, 2, 2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	compiler := New()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong, expected %d, actual %d",
			0, compiler.scopeIndex)
	}
	globalSymbolTable := compiler.symbolTable // ++

	compiler.emit(code.OpMul)

	// 进入一层 scope
	compiler.enterScope()
	if compiler.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong, expected %d, actual %d",
			1, compiler.scopeIndex)
	}

	compiler.emit(code.OpSub)
	if len(compiler.scopes[compiler.scopeIndex].instructions) != 1 {
		t.Errorf("instructions length wrong, expected 1, actual %d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	last := compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpSub {
		t.Errorf("lastInstruction.Opcode wrong, expected %d, actual %d",
			code.OpSub, last.Opcode)
	}

	// ++
	if compiler.symbolTable.Outer != globalSymbolTable {
		t.Errorf("compiler did not enclose symbolTable")
	}

	// 退出一层 scope
	compiler.leaveScope()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong, expected %d, actual %d",
			0, compiler.scopeIndex)
	}

	// ++
	if compiler.symbolTable != globalSymbolTable {
		t.Errorf("compiler did not restore global symbol table")
	}

	// ++
	if compiler.symbolTable.Outer != nil {
		t.Errorf("compiler modified global symbol table incorrectly")
	}

	compiler.emit(code.OpAdd)

	if len(compiler.scopes[compiler.scopeIndex].instructions) != 2 {
		t.Errorf("instructions length wrong, expected 2, actual %d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	// 检查 lastInstruction, previousInstruction
	last = compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInstruction.Opcode wrong, expected %d, actual %d",
			code.OpAdd, last.Opcode)
	}

	previous := compiler.scopes[compiler.scopeIndex].previousInstruction
	if previous.Opcode != code.OpMul {
		t.Errorf("previousInstruction.Opcode wrong, expected %d, actual %d",
			code.OpMul, previous.Opcode)
	}

}

func TestFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { return 5 + 10 }`,
			expectedConstants: []interface{}{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0), // 5
					code.Make(code.OpConstant, 1), // 10
					code.Make(code.OpAdd),         // ADD
					code.Make(code.OpReturnValue), // RETURN
				},
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 2), // --
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctions2(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { 5 + 10 }`,
			expectedConstants: []interface{}{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 2), // --
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { 1; 2 }`,
			expectedConstants: []interface{}{
				1,
				2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 2), // --
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { }`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 0), // --
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestLetStatementScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			let num = 55;
			fn() { num }
			`,
			expectedConstants: []interface{}{
				55,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				// code.Make(code.OpConstant, 1), // --
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			fn() {
				let num = 55;
				num
			}
			`,
			expectedConstants: []interface{}{
				55,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 1), // --
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			fn() {
				let a = 55;
				let b = 77;
				a + b
			}
			`,
			expectedConstants: []interface{}{
				55,
				77,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 2), // --
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { 24 }();`,
			expectedConstants: []interface{}{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0), // The literal "24"
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 1), // The compiled function
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpCall, 0), // ++ ",0"
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let noArg = fn() { 24 };
			noArg();
			`,
			expectedConstants: []interface{}{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0), // The literal "24"
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 1), // The compiled function
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0), // ++ ",0"
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionCalls2(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			let oneArg = fn(a) { };
			oneArg(24);
			`,
			expectedConstants: []interface{}{
				[]code.Instructions{ // `fn(a) {}`
					code.Make(code.OpReturn),
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 0),  // function literal
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0), // let oneArg = ...
				code.Make(code.OpGetGlobal, 0), // oneArg(...
				code.Make(code.OpConstant, 1),  // 1->24,
				code.Make(code.OpCall, 1),      // ...) call
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let manyArg = fn(a, b, c) { };
			manyArg(24, 25, 26);
			`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				24,
				25,
				26,
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 0),
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1), // 1->24
				code.Make(code.OpConstant, 2), // 2->25
				code.Make(code.OpConstant, 3), // 3->26
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionCalls3(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			let oneArg = fn(a) {a };
			oneArg(24);
			`,
			expectedConstants: []interface{}{
				[]code.Instructions{ // `fn(a) {}`
					code.Make(code.OpGetLocal, 0), // ++
					code.Make(code.OpReturnValue), // **
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 0),  // function literal
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0), // let oneArg = ...
				code.Make(code.OpGetGlobal, 0), // oneArg(...
				code.Make(code.OpConstant, 1),  // 1->24,
				code.Make(code.OpCall, 1),      // ...) call
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let manyArg = fn(a, b, c) { a; b; c };
			manyArg(24, 25, 26);
			`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0), // ++
					code.Make(code.OpPop),         // ++
					code.Make(code.OpGetLocal, 1), // ++
					code.Make(code.OpPop),         // ++
					code.Make(code.OpGetLocal, 2), // ++
					code.Make(code.OpReturnValue), // **
				},
				24,
				25,
				26,
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 0),
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1), // 1->24
				code.Make(code.OpConstant, 2), // 2->25
				code.Make(code.OpConstant, 3), // 3->26
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBuiltins(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			len([]);
			push([], 1);
			`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetBuiltin, 0), // 获取 0 号内置函数
				code.Make(code.OpArray, 0),      // 实参 0
				code.Make(code.OpCall, 1),       // call，实参数量为 1
				code.Make(code.OpPop),           // ;
				code.Make(code.OpGetBuiltin, 5), // 获取 5 号内置函数
				code.Make(code.OpArray, 0),      // 实参 0
				code.Make(code.OpConstant, 0),   // 实参 1
				code.Make(code.OpCall, 2),       // call，实参数量为 2
				code.Make(code.OpPop),           // ;
			},
		},
		{
			input: `fn() { len([]) }`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 0),
					code.Make(code.OpArray, 0),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				// code.Make(code.OpConstant, 0),
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			fn(a) {
				fn(b) {
					a + b
				}
			}
			`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),  // a
					code.Make(code.OpGetLocal, 0), // b
					code.Make(code.OpAdd),         // +
					code.Make(code.OpReturnValue), // return
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),   // a，来自 FreeSymbols 列表 ;; 压入闭包函数所需要的所有局部变量
					code.Make(code.OpClosure, 0, 1), // fn(b)...
					code.Make(code.OpReturnValue),   // return fn(b)...
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			fn(a) {
				fn(b) {
					fn(c) {
						a + b + c
					}
				}
			};
			`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),  // a
					code.Make(code.OpGetFree, 1),  // b
					code.Make(code.OpAdd),         // +
					code.Make(code.OpGetLocal, 0), // c
					code.Make(code.OpAdd),         // +
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),    // a，来自 FreeSymbols 列表
					code.Make(code.OpGetLocal, 0),   // b，来自 FreeSymbols 列表
					code.Make(code.OpClosure, 0, 2), // fn(c)..., 捕获 2 个局部变量
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),   // a，来自 FreeSymbols 列表
					code.Make(code.OpClosure, 1, 1), // fn(b)..., 捕获 1 个局部变量
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let global = 55;
			fn() {
				let a = 66;
				fn() {
					let b = 77;
					fn() {
						let c = 88;
						global + a + b + c;
					}
				}
			}
			`,
			expectedConstants: []interface{}{
				55,
				66,
				77,
				88,
				[]code.Instructions{
					code.Make(code.OpConstant, 3),  // 88
					code.Make(code.OpSetLocal, 0),  // let c=...
					code.Make(code.OpGetGlobal, 0), // 55 (global)
					code.Make(code.OpGetFree, 0),   // a->66
					code.Make(code.OpAdd),          //
					code.Make(code.OpGetFree, 1),   // b->77
					code.Make(code.OpAdd),          //
					code.Make(code.OpGetLocal, 0),  // c
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 2), // 77
					code.Make(code.OpSetLocal, 0), // let b=...
					code.Make(code.OpGetFree, 0),  // a，来自 FreeSymbols 列表
					code.Make(code.OpGetLocal, 0), // b
					code.Make(code.OpClosure, 4, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 5, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 6, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			let countDown = fn(x) { countDown(x - 1); };
			countDown(1);
			`,
			expectedConstants: []interface{}{
				1,
				[]code.Instructions{
					code.Make(code.OpCurrentClosure),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSub),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
				1,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let wrapper = fn() {
				let countDown = fn(x) { countDown(x - 1); };
				countDown(1);
			};
			wrapper();
			`,
			expectedConstants: []interface{}{
				1,
				[]code.Instructions{
					code.Make(code.OpCurrentClosure),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSub),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
				1,
				[]code.Instructions{
					code.Make(code.OpClosure, 1, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 2),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 3, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}
