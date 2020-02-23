package destination_test

import (
	"example.com/destination/testpkg"
	"example.com/epackage"
	"example.com/source/testpkg/nested"
)

func exampleFunc() {
	testpkg.ExampleFunc()
	nested.Stuff()
	epackage.Func()
}
