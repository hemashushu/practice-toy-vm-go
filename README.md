# (Practice) Toy VM - Go

<!-- @import "[TOC]" {cmd="toc" depthFrom=1 depthTo=6 orderedList=false} -->

<!-- code_chunk_output -->

- [(Practice) Toy VM - Go](#practice-toy-vm-go)
  - [使用方法](#使用方法)
    - [编译](#编译)
    - [进入 REPL 模式（交互模式）](#进入-repl-模式交互模式)
    - [运行指定的脚本](#运行指定的脚本)
    - [编译脚本并输出汇编文本](#编译脚本并输出汇编文本)
    - [运行脚本的示例](#运行脚本的示例)

<!-- /code_chunk_output -->

练习单纯使用 Go lang 编写简单的 _玩具VM_。

> 注：本项目是阅读和学习《Writing A Compiler In Go》时的随手练习，并无实际用途。程序的原理、讲解和代码的原始出处请移步 https://compilerbook.com/

## 使用方法

### 编译

`$ go build -o vm`

### 进入 REPL 模式（交互模式）

`$ ./vm`

或者

`$ go run .`

### 运行指定的脚本

`$ ./vm path_to_script_file`

或者

`$ go run . path_to_script_file`

### 编译脚本并输出汇编文本

`$ ./vm path_to_script_file -s`

或者

`$ go run . path_to_script_file -s`

### 运行脚本的示例

`$ ./toy examples/01-expression.toy`

或者

`$ go run . examples/01-expression.toy`

如无意外，应该能看到输出 `3`。
