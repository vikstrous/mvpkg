package destination

import (
	"example.com/source/testpkg"
)

type A struct {
	prop string
}

func exampleFunc3() {
	testpkg.ExampleFunc()
	testpkg := A{}
	testpkg.prop = "a"
}
