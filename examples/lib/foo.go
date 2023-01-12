package lib

import (
	"fmt"
)

type Foo interface {
	DoSomething()
}

type foo struct {
	name string
}

//@Component
func NewFoo() Foo {
	return &foo{}
}

func (f *foo) DoSomething() {
	println(f.name + " do something")
	fmt.Print("")
}
