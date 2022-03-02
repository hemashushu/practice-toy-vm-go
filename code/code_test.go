package code

import "testing"

// 测试 "编译"（将指令及其参数转为 byte 数组）
func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode // 操作码
		operands []int  // 参数（int 类型的数组）
		expected []byte // 预期产生的字节码的指令部分（byte 类型的数组）
	}{
		{
			OpConstant,                         // 操作码
			[]int{65534},                       // 参数
			[]byte{byte(OpConstant), 255, 254}, // 预期的字节码的指令部分
		},
		{
			OpAdd,
			[]int{},
			[]byte{byte(OpAdd)},
		},
		{
			OpGetLocal,
			[]int{255},
			[]byte{byte(OpGetLocal), 255},
		},
		{
			OpClosure,
			[]int{65534, 255},
			[]byte{byte(OpClosure), 255, 254, 255},
		},
	}

	for _, test := range tests {
		instruction := Make(test.op, test.operands...)

		if len(instruction) != len(test.expected) {
			t.Errorf("instruction length expected %d, actual %d",
				len(test.expected), len(instruction))
		}

		for idx, value := range test.expected {
			if instruction[idx] != value {
				t.Errorf("[%d] expected %d, actual %d", idx, value, instruction[idx])
			}
		}
	}
}

// 测试 ”反编译“（将字节码的指令部分，即一个 byte 数组，转为文本）
func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected :=
		`0000 OpConstant 1
0003 OpConstant 2
0006 OpConstant 65535
`

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}
	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted, expected %q, actual %q",
			expected, concatted.String())
	}
}

func TestInstructionsString2(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
`

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}
	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted, expected %q, actual %q",
			expected, concatted.String())
	}
}

func TestInstructionsString3(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpGetLocal, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected :=
		`0000 OpAdd
0001 OpGetLocal 1
0003 OpConstant 2
0006 OpConstant 65535
`
	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}
	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted, expected %q, actual %q",
			expected, concatted.String())
	}
}

func TestInstructionsString4(t *testing.T) {
	instructions := []Instructions{
		Make(OpConstant, 65535),
		Make(OpClosure, 65535, 255),
		Make(OpConstant, 2),
	}

	expected :=
		`0000 OpConstant 65535
0003 OpClosure 65535 255
0007 OpConstant 2
`
	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}
	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted, expected %q, actual %q",
			expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
		{OpGetLocal, []int{255}, 1},
		{OpClosure, []int{65535, 255}, 3},
	}
	for _, test := range tests {
		instruction := Make(test.op, test.operands...)
		def, err := Lookup(byte(test.op))
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}
		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != test.bytesRead {
			t.Fatalf("bytesRead wrong, expected %d, actual %d", test.bytesRead, n)

		}
		for i, want := range test.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong, expected %d, actual %d", want, operandsRead[i])
			}
		}
	}
}
