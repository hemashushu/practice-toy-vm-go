package vm

import (
	"fmt"
	"toyvm/code"
	"toyvm/compiler"
	"toyvm/object"
)

const StackSize = 2048    // 栈容量
const GlobalsSize = 65536 // 符号容量

var True = &object.Boolean{Value: true}   // Object 常量
var False = &object.Boolean{Value: false} // Object 常量
var Null = &object.Null{}                 // Object 常量

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object

	// Always points to the next value. Top of stack is stack[sp-1]
	// 栈指针，指向下一个空闲的位置。如果栈有一个元素，则 sp 的值为 1，所以也可以
	// 视为栈的当前元素的数量
	sp int

	globals []object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
		globals:      make([]object.Object, GlobalsSize),
	}
}

func NewWithGlobalsStore(
	bytecode *compiler.Bytecode,
	globals []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = globals
	return vm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		// fetch
		op := code.Opcode(vm.instructions[ip])

		// decode
		switch op {

		// 定义常量
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
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
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2 // 因为 OpJumpNotTruthy 指令一共 3 个字节，另外 for 循环会 +1，所以下一条指令的位置是 ip + 3 - 1

			condition := vm.pop()
			if !isTruthy(condition) {
				ip = pos - 1 // 因为 for 循环会 +1，所以 pos 需要 - 1
			}

		// 无条件跳转
		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1 // 因为 for 循环会 +1，所以 pos 需要 - 1

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
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

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
