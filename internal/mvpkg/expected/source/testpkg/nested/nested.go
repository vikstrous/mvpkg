package nested

import (
	"fmt"

	"example.com/destination/testpkg2"
)

func Stuff() {
	testpkg2.ExampleFunc()
	fmt.Println("stuff")
}
