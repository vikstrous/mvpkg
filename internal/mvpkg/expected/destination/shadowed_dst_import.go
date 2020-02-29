package destination

import (
	"example.com/destination/testpkg2"
	// this import will conflict with the new name
	testpkg2_ "example.com/source/testpkg2"
)

type A struct {
	prop string
}

func exampleFunc3() {
	testpkg2_.ExampleFunc()
	// a comment
	testpkg2.ExampleFunc()
}
