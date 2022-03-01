package compiler

import (
	"fmt"
	"toyvm/ast"
	"toyvm/code"
	"toyvm/object"
)

// 用于跟踪最后两个指令（名称及位置）
type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	instructions code.Instructions // 字节码的指令部分，[]byte
	constants    []object.Object   // 字节码的数据部分，[]Object

	lastInstruction     EmittedInstruction // 最近一次指令
	previousInstruction EmittedInstruction // 倒数第二次指令

	symbolTable *SymbolTable
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},

		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},

		symbolTable: NewSymbolTable(),
	}
}

func NewWithState(
	symbolTable *SymbolTable,
	constants []object.Object) *Compiler {

	compiler := New()
	compiler.symbolTable = symbolTable
	compiler.constants = constants
	return compiler
}

// "字节码" 包含了指令部分（.text）和数据部分(.data)
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
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

	case *ast.BlockStatement:
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

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// 使用一个临时的数值 `0` 作为 OpJumpNotTruthy 指令的参数
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 0)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		// Consequence 可能是一个语句块，假如最后的栈顶的值被（语句末尾的 OpPop 指令）移除，
		// 则移除 OpPop 指令
		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		// 为 consequence 段补上一个 OpJump 指令
		// 使用一个临时的数值 `0` 作为 OpJump 指令的参数
		jumpPos := c.emit(code.OpJump, 0)

		// 记录 alternative 语句块开始位置
		alternativePos := len(c.instructions)

		// 判断是否存在 Alternative

		if node.Alternative == nil { // 不存在 alternative
			// 因为缺少 alternative 指令，为了防止 if 的条件不成立时可以返回 Null 值，这里补上
			// OpNull 指令
			c.emit(code.OpNull)

		} else { // 存在 alternative
			// 生成 alternative 指令
			err = c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.instructions)

		// 更新临时参数
		c.changeOperand(jumpNotTruthyPos, alternativePos)
		c.changeOperand(jumpPos, afterAlternativePos)

	// 标识符定义和赋值语句
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)

	// 二元操作
	case *ast.InfixExpression:
		left, right, operator := node.Left, node.Right, node.Operator

		if operator == "<" {
			left = node.Right
			right = node.Left
			operator = ">"
		}

		err := c.Compile(left)
		if err != nil {
			return err
		}

		err = c.Compile(right)
		if err != nil {
			return err
		}

		switch operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)

		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case ">":
			c.emit(code.OpGreaterThan)

		default:
			return fmt.Errorf("unknown operator %s", operator)
		}

	// 一元操作
	case *ast.PrefixExpression:
		value := node.Right
		err := c.Compile(value)
		if err != nil {
			return err
		}

		operator := node.Operator
		switch operator {
		case "-":
			c.emit(code.OpMinus)
		case "!":
			c.emit(code.OpBang)
		default:
			return fmt.Errorf("unknown operator %s", operator)
		}

	// 标识符
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.emit(code.OpGetGlobal, symbol.Index)

	// 字面量
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
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

	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}

// 这个方法不能连续调用，因为 c.previousInstruction 无法恢复
func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	// 注：c.previousInstruction 没有被恢复
	c.lastInstruction = c.previousInstruction
}

// 替换长度相同的指令（[]byte）
func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

// 替换指定指令（[]byte）的参数（仅限一个参数，且参数长度需要相同）
func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}
