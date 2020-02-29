package destination

import (
	"example.com/destination/testpkg2"
	"example.com/destination/testpkg2/nested"
	"example.com/epackage"
)

func exampleFunc() {
	testpkg2.ExampleFunc()
	nested.Stuff()
	epackage.Func()
}
