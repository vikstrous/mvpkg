package destination

import (
	"example.com/destination/testpkg2"
)

type A struct {
	prop string
}

func exampleFunc3() {
	testpkg2_ := A{}
	testpkg2_.prop = "a"
	testpkg2.ExampleFunc()
}
