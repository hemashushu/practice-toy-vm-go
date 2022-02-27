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
			OpConstant, []int{65534}, // 操作码和参数
			[]byte{byte(OpConstant), 255, 254}, // 预期的字节码的指令部分
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

	expected := `0000 OpConstant 1
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

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}
	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)
		def, err := Lookup(byte(tt.op))
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}
		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != tt.bytesRead {
			t.Fatalf("n wrong. want=%d, got=%d", tt.bytesRead, n)

		}
		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}
