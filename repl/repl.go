package repl

import (
	"bufio"
	"fmt"
	"io"
	"toyvm/compiler"
	"toyvm/lexer"
	"toyvm/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
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

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compilation failed: %s\n", err)
			continue
		}

		fmt.Println(comp.Bytecode().Instructions.String())

		// 		machine := vm.New(comp.Bytecode())
		// 		err = machine.Run()
		// 		if err != nil {
		// 			fmt.Fprintf(out, "Executing bytecode failed: %s\n", err)
		// 			continue
		// 		}
		//
		// 		// stackTop := machine.StackTop()
		// 		lastPopped := machine.LastPoppedStackElem()
		// 		io.WriteString(out, lastPopped.Inspect())
		// 		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
