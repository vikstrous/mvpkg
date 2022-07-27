package depender

import "example.com/destination/target"

type Bar struct {
	target target.Foo
}

func (b *Bar) Foo() target.Foo {
	b.target.Hello()
	return target.Foo{}
}
