package destination_test

import (
	"example.com/destination/testpkg2"
	"example.com/epackage"
	"example.com/source/testpkg/nested"
)

func exampleFunc() {
	testpkg2.ExampleFunc()
	nested.Stuff()
	epackage.Func()
}
