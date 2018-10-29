package monkey

import (
	"fmt"
)

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

func isAlreadyReplaced(bytes []byte) bool {
	return bytes[0] == 0x48 && bytes[1] == 0xBA && bytes[10] == 0xFF && bytes[11] == 0x22
}

func codeOffset(from uintptr) uintptr {
	//
	// 直接找sub指令的特征，有可能会有例外目测撞上的概率不高，等发现例外了再加入更多的特征
	//
	var inArray = func(v byte, arr []byte) bool {
		for _, vv := range arr {
			if v == vv {
				return true
			}
		}
		return false
	}

	mid := []byte{0x81, 0x83, 0x8d}
	f := rawMemoryAccess(from, 40)
	for _, v := range []int{15, 19, 24, 27, 31} {
		if f[v] == byte(0x48) && f[v+2] == byte(0xec) && inArray(f[v+1], mid) {
			return uintptr(v)
		}
	}

	panic("\nraw data: \n")
	var i = 0
	for _, v := range f {
		i++
		fmt.Printf("0x%02X ", v)
		if i > 0 && (i%8) == 0 {
			fmt.Printf("\n")
		}
	}
	panic("\noffset not fund\n")

	return 0
}
