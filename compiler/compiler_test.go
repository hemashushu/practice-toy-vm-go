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

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
			},
		},
	}

	runCompilerTests(t, tests)
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
		return fmt.Errorf("instructions length expected %q, actual %q",
			concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("[%d] instruction expected %q, actual %q",
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
				return fmt.Errorf("constant %d - testIntegerObject failed: %s",
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
