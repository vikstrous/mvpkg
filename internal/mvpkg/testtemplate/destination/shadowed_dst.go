package destination

import (
	"example.com/source/testpkg"
)

type A struct {
	prop string
}

func exampleFunc3() {
	testpkg2 := A{}
	testpkg2.prop = "a"
	testpkg.ExampleFunc()
}
