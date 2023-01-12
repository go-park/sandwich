package lib

type Foo interface {
	DoSomething()
}

type foo struct {
	name string
}

func NewBar() Foo {
	return &foo{}
}

func (f *foo) DoSomething() {
	println(f.name + " do something")
}
