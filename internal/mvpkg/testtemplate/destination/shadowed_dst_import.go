package destination

import (
	"example.com/source/testpkg"
	// this import will conflict with the new name
	"example.com/source/testpkg2"
)

type A struct {
	prop string
}

func exampleFunc3() {
	testpkg2.ExampleFunc()
	// a comment
	testpkg.ExampleFunc()
}
