package main

import (
	"fmt"
	"os"
	"toyvm/repl"
)

func main() {
	args := os.Args
	count := len(args)

	if count == 1 {
		// 进入 REPL 交互模式
		fmt.Println("Toy lang REPL")
		repl.Start(os.Stdin, os.Stdout)

	} else if count == 2 {
		// 编译及执行脚本
		// executor.Exec(args[1])

	} else {
		fmt.Println(`Toy VM interpreter
Usage:

1. Launch REPL mode
$ go run .

2. Execute toy lang script source code file
$ go run . path_to_script_file`)
	}
}
