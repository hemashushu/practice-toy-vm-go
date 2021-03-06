package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type Opcode byte

// 操作码（指令）详细信息，可视为 "函数的签名" 信息。
type Definition struct {
	Name          string // 指令名称
	OperandWidths []int  // 参数的个数，以及各个参数的长度（单位为 byte）
}

// 操作码（指令）列表，可视为 "函数名称列表"。
const (
	OpConstant Opcode = iota // 从 global 读取常量，并压入运算栈
	OpPop                    // 弹出语句最后的值

	OpAdd // 加
	OpSub // 减
	OpMul // 乘
	OpDiv // 除

	OpTrue  // 向栈压入 True
	OpFalse // 向栈压入 False
	OpNull  // 向栈压入 Null

	OpEqual       // ==
	OpNotEqual    // !=
	OpGreaterThan // >

	OpMinus // -
	OpBang  // !

	OpJumpNotTruthy // 非 true 时跳转
	OpJump          // 无条件跳转

	OpGetGlobal // 获取标识符的值
	OpSetGlobal // 保存标识符的值

	// 原书的实践是支持 String, Array, Map 等复杂数据类型，实际上可能
	// 单单支持数字（包括 Integer, Boolean, Null）会比较简单，至于
	// 复杂数据类型可以通过标准库实现。
	OpArray // 生成 Array
	OpHash  // 生成 Hash（Map）
	OpIndex // Array 和 Hash 的索引访问

	OpCall        // 调用函数
	OpReturnValue // 从函数返回，返回一个值
	OpReturn      // 从函数返回，无返回值

	OpGetLocal // 读局部变量（包括函数参数）
	OpSetLocal // 写局部变量

	OpGetBuiltin // 获取内置函数

	OpClosure // 创建闭包
	OpGetFree // 读取闭包中捕获的局部变量的值

	OpCurrentClosure
)

// 操作码（指令）详细信息列表
var definitions = map[Opcode]*Definition{
	// OpConstant
	// 作用：从 global 读取常量，并压入运算栈
	// 参数：1. UInt16，记录数值在常量列表中的地址
	OpConstant: {"OpConstant", []int{2}},

	// OpPop
	// 作用：弹出语句最后的值
	// 参数：无
	OpPop: {"OpPop", []int{}},

	// OpAdd
	// 作用：两个数相加
	// 参数：无
	OpAdd: {"OpAdd", []int{}},
	OpSub: {"OpSub", []int{}},
	OpMul: {"OpMul", []int{}},
	OpDiv: {"OpDiv", []int{}},

	// OpTrue/OpFalse
	// 作用：向 stack 压入 True 或者 False
	// 参数：无
	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},
	OpNull:  {"OpNull", []int{}},

	// OpEqual/OpNotEqual/OpGreaterThan
	// 比较运算
	OpEqual:       {"OpEqual", []int{}},
	OpNotEqual:    {"OpNotEqual", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},

	// OpMinus/OpBang
	// 一元操作
	OpMinus: {"OpMinus", []int{}},
	OpBang:  {"OpBang", []int{}},

	// 非 true 时跳转
	// 用于跳到 else 语句块开始位置，或者
	// 跳到 if 语句整体结束的位置（在缺少 else 语句块时）
	// 参数：1. UInt16 目标位置
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},

	// 无条件跳转
	// 跳到 if 语句整体结束的位置
	// 参数：1. UInt16 目标位置
	OpJump: {"OpJump", []int{2}},

	// 获取标识符的值
	// 从 global 列表里获取指定位置的值，然后压入栈
	// 参数：1. UInt16 目标位置
	OpGetGlobal: {"OpGetGlobal", []int{2}},

	// 保存标识符的值
	// 弹出栈顶的值，并把值保存到 global 列表
	// 参数：1. UInt16 目标在 global 列表里的位置
	OpSetGlobal: {"OpSetGlobal", []int{2}},

	// 使用栈的前 N 项数值生成一个 Array 对象
	// 参数：1. UInt16 元素的个数
	OpArray: {"OpArray", []int{2}},

	// 使用栈的前 N 项键值对生成一个 Hash(Map) 对象
	// 参数：1. UInt16 键值对（Pair）的个数
	OpHash: {"OpHash", []int{2}},

	// Array 和 Hash 的索引访问
	// 栈顶是 "索引值"，栈倒数第二个数是 "Array 或 Map 对象"
	OpIndex: {"OpIndex", []int{}},

	// 调用位于栈顶的 object.CompiledFunction
	// 参数：1. UInt8 调用函数时实参的数量
	OpCall: {"OpCall", []int{1}},

	// 返回栈顶的一个数值
	OpReturnValue: {"OpReturnValue", []int{}},

	// 返回 vm.Null
	OpReturn: {"OpReturn", []int{}},

	// 读写局部变量
	// 参数：1. UInt8 目标在运算栈中的位置
	OpGetLocal: {"OpGetLocal", []int{1}},
	OpSetLocal: {"OpSetLocal", []int{1}},

	// 获取内置函数
	// 参数：1. UInt8 内置函数的索引值
	OpGetBuiltin: {"OpGetBuiltin", []int{1}},

	// 创建闭包
	// 参数：1. UInt16 constant index，指向目标 *object.CompiledFunction
	// 参数：1. UInt8 闭包局部变量的数量
	OpClosure: {"OpClosure", []int{2, 1}},

	OpGetFree: {"OpGetFree", []int{1}},

	OpCurrentClosure: {"OpCurrentClosure", []int{}},
}

// 编译
// 将 "操作码及其参数" 转换为字节数组
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1 // 操作码（Opcode）本身占用一个 byte
	for _, width := range def.OperandWidths {
		instructionLen += width
	}

	instruction := make([]byte, instructionLen) // 构建一个 byte 数组
	instruction[0] = byte(op)                   // 填充第一个 byte

	offset := 1
	for idx, operand := range operands {
		width := def.OperandWidths[idx]

		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(operand))
		case 1:
			instruction[offset] = byte(operand)
		}

		offset += width
	}

	return instruction
}

// 反编译
// 将字节码（包含有一个或多个指令）当中的指令部分（一个 byte 数组）转换为字符串
// e.g.
// "0000 OpConstant 1"
// "0003 OpConstant 2"
// "0006 OpConstant 65535"
func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])

		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += 1 + read
	}
	return out.String()
}

// 反编译
// 格式化指令名称、参数值
func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return fmt.Sprintf("%s", def.Name)
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// 反编译
// 将字节码当中————指令部分当中的————参数部分（一个 byte 数组） "还原" 为
// 操作码的参数（Operands）列表（一个 int 数组）
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

// 根据操作码（指令）查找操作码详细信息
// 当前仅用于测试
// 在 VM 的执行过程中，为了效率而直接硬编码操作码的详细信息
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}
