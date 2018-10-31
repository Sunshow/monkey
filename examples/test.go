package main

import (
	"fmt"

	"github.com/sychenjianjun/monkey"
)

func innerCall(param1, param2, param3, param4 [2000]byte) bool {
	fmt.Println("aaa", len(param1), len(param2), len(param3), len(param4))
	return false
}

// 测试函数
func myfunc() (ret int) {
	fmt.Println("func origin")
	ret = 1

	var v = [2000]byte{3, 4}
	innerCall(v, v, v, v)

	return
}

func main() {
	fmt.Println("test func")
	monkey.PatchEx(myfunc, originmyfunc, func() (ret int) {
		fmt.Println("func mocked")
		originmyfunc()
		return
	})

	myfunc()

	monkey.Unpatch(myfunc)
	myfunc()
	originmyfunc()
}

func originmyfunc() (ret int) {
	fmt.Println("")
	return 0
}
