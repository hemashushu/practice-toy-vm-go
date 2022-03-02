package vm

import (
	"fmt"
	"toyvm/code"
	"toyvm/compiler"
	"toyvm/object"
)

const StackSize = 2048    // 运算栈容量
const GlobalsSize = 65536 // 符号容量
const MaxFrames = 1024    // 调用栈的容量

var True = &object.Boolean{Value: true}   // Object 常量
var False = &object.Boolean{Value: false} // Object 常量
var Null = &object.Null{}                 // Object 常量

type VM struct {
	constants []object.Object
	// instructions code.Instructions

	// 运算栈
	// 注意运算栈里存储的不是单纯数字（比如 int），而是 object.Object 对象
	stack []object.Object

	// Always points to the next value. Top of stack is stack[sp-1]
	// 栈指针，指向下一个空闲的位置。如果栈有一个元素，则 sp 的值为 1，所以也可以
	// 视为栈当中当前元素的数量，准确名称是 stackCount
	sp int

	// 全局标识符
	globals []object.Object

	frames     []*Frame // 调用帧列表
	frameIndex int      // 调用帧的数量，准确名称是 frameCount
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{
		Instructions: bytecode.Instructions,
	}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		// instructions: bytecode.Instructions,

		constants: bytecode.Constants,
		stack:     make([]object.Object, StackSize),
		sp:        0, // 栈当中当前元素的数量，准确名称是 stackCount
		globals:   make([]object.Object, GlobalsSize),

		frames:     frames,
		frameIndex: 1, // 调用帧的数量，准确名称是 frameCount
	}
}

func NewWithGlobalsStore(
	bytecode *compiler.Bytecode,
	globals []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = globals
	return vm
}

// func (vm *VM) StackTop() object.Object {
// 	if vm.sp == 0 {
// 		return nil
// 	}
//
// 	return vm.stack[vm.sp-1]
// }

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	// for ip := 0; ip < len(vm.instructions); ip++ {
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		// fetch
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])
		// op := code.Opcode(vm.instructions[ip])

		// decode
		switch op {

		// 定义常量
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:]) // code.ReadUint16(vm.instructions[ip+1:])
			vm.currentFrame().ip += 2                 // ip += 2

			// execute
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		// 弹出栈顶的最后一个值，用于清理语句执行后的 stack
		case code.OpPop:
			vm.pop()

		// 条件跳转（false 时跳转）
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:])) // int(code.ReadUint16(vm.instructions[ip+1:]))
			// ip += 2                                 // 因为 OpJumpNotTruthy 指令一共 3 个字节，另外 for 循环会 +1，所以下一条指令的位置是 ip + 3 - 1
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				// ip = pos - 1 // 因为 for 循环会 +1，所以 pos 需要 - 1
				vm.currentFrame().ip = pos - 1
			}

		// 无条件跳转
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:])) // int(code.ReadUint16(vm.instructions[ip+1:]))
			// ip = pos - 1                            // 因为 for 循环会 +1，所以 pos 需要 - 1
			vm.currentFrame().ip = pos - 1

		// 函数调用
		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:]) // 参数的数量
			vm.currentFrame().ip += 1

			err := vm.callFunction(int(numArgs))
			if err != nil {
				return err
			}

		case code.OpReturnValue:
			returnValue := vm.pop()

			// vm.popFrame()
			// vm.pop()
			frame := vm.popFrame()

			// 重置 sp 为 frame.basePointer，用于清除保留局部变量空间
			// `- 1` 相当于 pop() 了一次
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}

		case code.OpReturn:
			vm.popFrame()
			vm.pop()

			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop() // 通过 “帧指针+偏移值” 计算出局部变量的位置

		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}

		// 加减乘除运算
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		// 标识符操作
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:]) // code.ReadUint16(vm.instructions[ip+1:])
			// ip += 2
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:]) // code.ReadUint16(vm.instructions[ip+1:])
			// ip += 2
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		// 一元操作
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		// 创建 Array
		case code.OpArray:
			count := int(code.ReadUint16(ins[ip+1:])) // int(code.ReadUint16(vm.instructions[ip+1:]))
			// ip += 2
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-count, vm.sp)

			// 改变 stack point 的值，相当于弹出 count 项数值
			vm.sp = vm.sp - count

			err := vm.push(array)

			if err != nil {
				return err
			}

		// 创建 Hash(Map)
		case code.OpHash:
			count := int(code.ReadUint16(ins[ip+1:])) // int(code.ReadUint16(vm.instructions[ip+1:]))
			// ip += 2
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-count, vm.sp)
			if err != nil {
				return err
			}

			// 改变 stack point 的值，相当于弹出 count 项数值
			vm.sp = vm.sp - count

			err = vm.push(hash)
			if err != nil {
				return err
			}

		// 读取索引
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		// 置布尔值操作
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}

		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop() // pop 的顺序应该与 push 的相反
	left := vm.pop()

	rightType := right.Type()
	leftType := left.Type()

	// if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
	// 	return vm.executeBinaryIntegerOperation(op, left, right)
	// }
	//
	// return fmt.Errorf("unsupported types for binary operation: %s %s",
	// 	leftType, rightType)

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)

	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)

	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s",
			leftType, rightType)
	}

}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode,
	left object.Object, right object.Object) error {

	rightValue := right.(*object.Integer).Value
	leftValue := left.(*object.Integer).Value

	//var result int64

	switch op {
	case code.OpAdd:
		// result = leftValue + rightValue
		return vm.push(&object.Integer{Value: leftValue + rightValue})
	case code.OpSub:
		// result = leftValue - rightValue
		return vm.push(&object.Integer{Value: leftValue - rightValue})
	case code.OpMul:
		// result = leftValue * rightValue
		return vm.push(&object.Integer{Value: leftValue * rightValue})
	case code.OpDiv:
		// result = leftValue / rightValue
		return vm.push(&object.Integer{Value: leftValue / rightValue})

	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	// return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode,
	left object.Object, right object.Object) error {

	// String 只支持 "+" 运算
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value
	return vm.push(&object.String{Value: leftValue + rightValue})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(
	op code.Opcode, left, right object.Object) error {

	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	} else {
		return False
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		// Null 视为 false
		// 注意 toy 语句没有 null 字面量
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()
	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		// 非 Boolean 和 Null 的数据都作为 true
		// 注意 toy 语句没有 null 字面量
		return true
	}
}

func (vm *VM) buildArray(start int, end int) object.Object {
	elements := make([]object.Object, end-start)

	for i := start; i < end; i++ {
		elements[i-start] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(start int, end int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := start; i < end; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable) // 判断 key 的数据类型是否 Hashable
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}
	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) executeIndexExpression(left object.Object, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)
	if i < 0 || i > max {
		return vm.push(Null)
	}
	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}
	return vm.push(pair.Value)
}

func (vm *VM) callFunction(numArgs int) error {
	// 注：
	// 调用帧从 `vm.sp` 开始
	// 当函数有参数时，运算帧的前 numArgs 个值都是实参
	// 所以调用帧的开始位置修正为 `vm.sp - numArgs`
	// CompiledFunction 的位置则是 `vm.sp - numArgs - 1`
	fn, ok := vm.stack[vm.sp-numArgs-1].(*object.CompiledFunction) // **
	if !ok {
		return fmt.Errorf("calling non-function")
	}

	// 检查实参的数量
	// 注：
	// 也可以在编译阶段检查
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments, expected %d, actual %d",
			fn.NumParameters, numArgs)
	}

	frame := NewFrame(fn, vm.sp-numArgs)
	vm.pushFrame(frame)                      // 压入新的调用帧
	vm.sp = frame.basePointer + fn.NumLocals // 保留空间给（自定义函数的）局部变量
	return nil
}
