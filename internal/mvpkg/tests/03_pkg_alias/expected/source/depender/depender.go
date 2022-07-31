package depender

import (
	target "example.com/destination/targetnew"
)

func Bar() {
	target.Foo()
}
