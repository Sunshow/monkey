package monkey

import "bytes"

// Assembles a jump to a function value
func jmpToFunctionValue(to uintptr) []byte {
	return []byte{
		0x48, 0xBA,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56), // movabs rdx,to
		0xFF, 0x22,     // jmp QWORD PTR [rdx]
	}
}

func checkFuncEmpty(data []byte, sz int) {
	// 检查一下是不是空函数
	int3 := []byte{0xcc, 0xcc, 0xcc, 0xcc, 0xcc}
	lenH := sz - 5
	for i := 0; i < lenH; i++ {
		if bytes.Equal(data[i:i+5], int3) {
			panic("alias不允许使用空函数\n")
		}
	}
}
