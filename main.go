package main

import (
	"fmt"
	"os"
	"toyvm/executor"
	"toyvm/repl"
)

func main() {
	args := os.Args
	count := len(args)

	if count == 1 {
		// 进入 REPL 交互模式
		fmt.Println("Toy VM REPL")
		repl.Start(os.Stdin, os.Stdout)

	} else if count == 2 {
		// 编译及执行脚本
		executor.Exec(args[1])

	} else if count == 3 && args[2] == "-s" {
		// 编译及打印汇编文本
		executor.Assembly(args[1])

	} else {
		fmt.Println(`Toy VM interpreter
Usage:

1. Launch REPL mode
$ go run .

2. Compile and execute toy lang script source code file
$ go run . path_to_script_file

3. Compile and print the assembly text
$ go run . path_to_script_file -s`)
	}
}
