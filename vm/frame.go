package vm

import (
	"toyvm/code"
	"toyvm/object"
)

// 调用帧/栈帧
// call frame, stack frame
// 对应着解析器的函数 activation record（在解析器里一般使用 Environment 实现）
type Frame struct {
	fn          *object.CompiledFunction
	ip          int
	basePointer int // BP/帧指针，进入调用帧之前，运算栈的栈顶位置（指针）
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn:          fn,
		ip:          -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
