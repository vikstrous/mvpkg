package nested

import (
	"fmt"

	"example.com/source/testpkg"
)

func Stuff() {
	testpkg.ExampleFunc()
	fmt.Println("stuff")
}
