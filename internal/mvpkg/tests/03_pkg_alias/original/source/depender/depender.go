package depender

import (
	target "example.com/source/target"
)

func Bar() {
	target.Foo()
}
