package main

import (
	"context"

	"github.com/go-park/sandwich/examples/lib"
	"gorm.io/gorm"
)

var _ IFoo = &Foo{}

//@Proxy("IFoo",option="FooOption", singleton=true)
type Foo struct {
	//@Inject
	foo lib.Foo
	//@Value("123")
	str string
	//@Value("true")
	boo bool
	//@Value("123")
	num uint64
}

type IFoo interface {
	Foo(ctx context.Context, i any, tx *gorm.DB) (any, error)
}

//@Pointcut("log", "trans")
func (s *Foo) Foo(ctx context.Context, i any, tx *gorm.DB) (any, error) {
	println("foo")
	return nil, nil
}
