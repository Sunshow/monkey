package main

import (
	"fmt"
	"time"

	"bou.ke/monkey"
)

func main() {
	fmt.Println(time.Now())

	monkey.PatchEx(time.Now, originNow, func() time.Time {
		return time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	})
	fmt.Println(time.Now())
	monkey.Unpatch(time.Now)
	fmt.Println(time.Now())
}

func originNow() time.Time {
	fmt.Println("nothing")
	return time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
}
