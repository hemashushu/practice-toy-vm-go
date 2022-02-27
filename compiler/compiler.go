package compiler

import (
	"fmt"
	"toyvm/ast"
	"toyvm/code"
	"toyvm/object"
)

type Compiler struct {
	instructions code.Instructions // 字节码的指令部分，[]byte
	constants    []object.Object   // 字节码的数据部分，[]Object
}

// "字节码" 包含了指令部分（.text）和数据部分(.data)
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

// 编译程序
// 结果是字节码，字节码包括指令部分和数据部分
func (c *Compiler) Compile(n ast.Node) error {
	switch node := n.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop) // 弹出语句的最后一个结果（用于清除栈）

	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	}

	return nil
}

// 将常量/字面量添加到常量列表，返回该常量的位置值
func (c *Compiler) addConstant(obj object.Object) int {
	idx := len(c.constants)
	c.constants = append(c.constants, obj)
	return idx
}

// 生成指令字节，返回该指令在字节中的位置值
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}
