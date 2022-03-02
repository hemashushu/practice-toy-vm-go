package compiler

import (
	"fmt"
	"sort"
	"toyvm/ast"
	"toyvm/code"
	"toyvm/object"
)

// 用于跟踪最后两个指令（名称及位置）
type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions        code.Instructions  // 字节码的指令部分，[]byte
	lastInstruction     EmittedInstruction // 最近一次指令
	previousInstruction EmittedInstruction // 倒数第二次指令
}

type Compiler struct {
	constants   []object.Object // 字节码的数据部分，[]Object
	symbolTable *SymbolTable

	// instructions        code.Instructions  // 字节码的指令部分，[]byte
	// lastInstruction     EmittedInstruction // 最近一次指令
	// previousInstruction EmittedInstruction // 倒数第二次指令

	scopes     []CompilationScope
	scopeIndex int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: NewSymbolTable(),

		// instructions:        code.Instructions{},
		// lastInstruction:     EmittedInstruction{},
		// previousInstruction: EmittedInstruction{},

		scopes:     []CompilationScope{mainScope},
		scopeIndex: 0,
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

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

// "字节码" 包含了指令部分（.text）和数据部分(.data)
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(), // c.instructions,
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
		// alternativePos := len(c.instructions)
		alternativePos := len(c.currentInstructions())

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

		// afterAlternativePos := len(c.instructions)
		afterAlternativePos := len(c.currentInstructions())

		// 更新临时参数
		c.changeOperand(jumpNotTruthyPos, alternativePos)
		c.changeOperand(jumpPos, afterAlternativePos)

	// 用户自定义函数
	case *ast.FunctionLiteral:
		c.enterScope()

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.replaceLastPopWithReturn()
		}

		// 针对函数体为空的用户自定义函数，或者
		// 最后一句为无返回值的语句，比如 `let` 语句。
		// 注：
		// 在书中的实践，如果函数无返回值，不是返回 Null，而是
		// 使用单独的一个指令 OpReturn。
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		numLocals := c.symbolTable.numDefinitions // 计算函数主体内的局部变量的数量
		instructions := c.leaveScope()

		compiledFn := &object.CompiledFunction{
			Instructions: instructions,
			NumLocals:    numLocals,
		}

		// 注：
		// 书中的实践是把用户自定义函数的指令（[]byte）当作一个 object.CompiledFunction
		// 对象存储在 c.constants 里，而不是合并到 instructions 里。
		// 这么做主要是为了简化实现的方法，不过一般的实践是合并到 instructions 里。
		c.emit(code.OpConstant, c.addConstant(compiledFn))

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		c.emit(code.OpCall)

	// 标识符定义和赋值语句
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		symbol := c.symbolTable.Define(node.Name.Value)

		if symbol.Scope == GlobalScope { // ++
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index) // ++
		}

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

	// 复合数据类型
	case *ast.ArrayLiteral:
		for _, element := range node.Elements {
			err := c.Compile(element)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		keys := []ast.Expression{} // 先获取 Hash（Map）Literal 的 keys
		for key := range node.Pairs {
			keys = append(keys, key)
		}

		// 对 keys 排序（可选的，主要为了方便测试，否则 key 的顺序是随机的）
		sort.Slice(keys, func(i int, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, key := range keys {
			err := c.Compile(key)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[key])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Pairs)*2)

	// 求 Array 和 Hash 的索引
	case *ast.IndexExpression:
		err := c.Compile(node.Left)

		if err != nil {
			return err
		}

		err = c.Compile(node.Index)

		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	// 标识符
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		if symbol.Scope == GlobalScope { // ++
			c.emit(code.OpGetGlobal, symbol.Index)
		} else {
			c.emit(code.OpGetLocal, symbol.Index) // ++
		}

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

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
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
	// posNewInstruction := len(c.instructions)
	// c.instructions = append(c.instructions, ins...)

	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	// previous := c.lastInstruction
	// last := EmittedInstruction{Opcode: op, Position: pos}
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	// c.previousInstruction = previous
	// c.lastInstruction = last
	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIsPop() bool {
	// return c.lastInstruction.Opcode == code.OpPop
	// return c.scopes[c.scopeIndex].lastInstruction.Opcode == code.OpPop
	return c.lastInstructionIs(code.OpPop)
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

// 这个方法不能连续调用，因为 c.previousInstruction 无法恢复
func (c *Compiler) removeLastPop() {
	// 注：c.previousInstruction 没有被恢复
	// c.instructions = c.instructions[:c.lastInstruction.Position]
	// c.lastInstruction = c.previousInstruction

	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	c.scopes[c.scopeIndex].instructions = c.currentInstructions()[:last.Position]
	c.scopes[c.scopeIndex].lastInstruction = previous
}

// 替换长度相同的指令（[]byte）
func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	// for i := 0; i < len(newInstruction); i++ {
	// 	c.instructions[pos+i] = newInstruction[i]
	// }

	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

// 替换指定指令（[]byte）的参数（仅限一个参数，且参数长度需要相同）
func (c *Compiler) changeOperand(opPos int, operand int) {
	// op := code.Opcode(c.instructions[opPos])
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

// 用于实现用户自定义函数的隠式 return
func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

// 压入一层 scope
func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// 弹出一层 scope，返回该层的指令（[]byte）
func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.Outer

	return instructions
}
