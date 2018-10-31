package monkey

import (
	"syscall"
	"unsafe"
)

const PAGE_EXECUTE_READWRITE = 0x40

var procVirtualProtect = syscall.NewLazyDLL("kernel32.dll").NewProc("VirtualProtect")

func virtualProtect(lpAddress uintptr, dwSize int, flNewProtect uint32, lpflOldProtect unsafe.Pointer) error {
	ret, _, _ := procVirtualProtect.Call(
		lpAddress,
		uintptr(dwSize),
		uintptr(flNewProtect),
		uintptr(lpflOldProtect))
	if ret == 0 {
		return syscall.GetLastError()
	}
	return nil
}

// this function is super unsafe
// aww yeah
// It copies a slice to a raw memory location, disabling all memory protection before doing so.
func copyToLocation(location uintptr, data []byte) {
	f := rawMemoryAccess(location, len(data))

	var oldPerms uint32
	err := virtualProtect(location, len(data), PAGE_EXECUTE_READWRITE, unsafe.Pointer(&oldPerms))
	if err != nil {
		panic(err)
	}
	copy(f, data[:])

	// VirtualProtect requires you to pass in a pointer which it can write the
	// current memory protection permissions to, even if you don't want them.
	var tmp uint32
	err = virtualProtect(location, len(data), oldPerms, unsafe.Pointer(&tmp))
	if err != nil {
		panic(err)
	}
}

//
// 把函数头第一条调栈的sub指令位置找出来，并返回函数头到这条指令之间的所有代码
//
func copyMoreStack(head []byte) []byte {
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
	for _, v := range []int{22, 26, 56} {
		if head[v] == byte(0x48) && head[v+2] == byte(0xec) && inArray(head[v+1], mid) {
			return head[0:v]
		}
	}

	printRawData(head)
	panic("offset not fund\n")

	return []byte{}
}
