package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type Opcode byte

// 操作码（指令）列表，可视为 "函数名称列表"。
const (
	OpConstant Opcode = iota // OpConstant 定义常量
)

// 操作码（指令）详细信息，可视为 "函数的签名" 信息。
type Definition struct {
	Name          string
	OperandWidths []int
}

// 操作码（指令）详细信息列表
var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
}

// 根据操作码（指令）查找操作码详细信息
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
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
		}

		offset += width
	}

	return instruction
}

// 反编译
// 将字节码当中的指令部分（一个 byte 数组）转换为字符串
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

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// 将字节码当中指令部分当中的参数部分（一个 byte 数组）
// 将字节数组 "还原" 为操作码的参数列表（即 Operands）
// 这时一个跟 Make 函数相反的操作
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
