package repl

import (
	"bufio"
	"fmt"
	"io"
	"toyvm/compiler"
	"toyvm/lexer"
	"toyvm/object"
	"toyvm/parser"
	"toyvm/vm"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	// 编译器和 VM 的状态
	constants := []object.Object{}
	symbolTable := compiler.NewSymbolTable()
	globals := make([]object.Object, vm.GlobalsSize)

	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)

		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compilation failed: %s\n", err)
			continue
		}

		machine := vm.NewWithGlobalsStore(comp.Bytecode(), globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Executing bytecode failed: %s\n", err)
			continue
		}

		// stackTop := machine.StackTop()
		lastPopped := machine.LastPoppedStackElem()
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
