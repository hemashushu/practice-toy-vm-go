package executor

import (
	"fmt"
	"os"
	"toyvm/compiler"
	"toyvm/lexer"
	"toyvm/parser"
	"toyvm/vm"
)

func Exec(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Read file error: %s\n", err)
		return
	}

	text := string(content)

	l := lexer.New(text)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(p.Errors())
		return
	}

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation failed: %s\n", err)
		return
	}

	machine := vm.New(comp.Bytecode())
	err = machine.Run()
	if err != nil {
		fmt.Printf("Executing bytecode failed: %s\n", err)
		return
	}

	// stackTop := machine.StackTop()
	lastPopped := machine.LastPoppedStackElem()
	fmt.Println(lastPopped.Inspect())
}

func Assembly(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Read file error: %s\n", err)
		return
	}

	text := string(content)

	l := lexer.New(text)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(p.Errors())
		return
	}

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation failed: %s\n", err)
		return
	}

	fmt.Println(comp.Bytecode().Instructions.String())
}

func printParserErrors(errors []string) {
	fmt.Println("Parser errors:")
	for _, msg := range errors {
		fmt.Println("\t" + msg)
	}
}
