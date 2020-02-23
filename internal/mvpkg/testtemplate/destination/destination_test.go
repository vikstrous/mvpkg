package destination

import (
	"example.com/epackage"
	"example.com/source/testpkg"
	"example.com/source/testpkg/nested"
)

func exampleFunc() {
	testpkg.ExampleFunc()
	nested.Stuff()
	epackage.Func()
}
