package destination

import (
	"example.com/destination/testpkg2"
)

type A struct {
	prop string
}

func exampleFunc3() {
	testpkg2.ExampleFunc()
	testpkg := A{}
	testpkg.prop = "a"
}
