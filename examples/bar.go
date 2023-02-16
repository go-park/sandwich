package main

import (
	"context"

	"github.com/go-park/sandwich/examples/lib"
	"gorm.io/gorm"
)

var _ IBar = &Bar{}

//@Proxy("IBar")
type Bar struct {
	//@Inject
	foo IFoo
	//@Inject
	libFoo lib.Foo
}
type IBar interface {
	Foo(ctx context.Context, i any, tx *gorm.DB) (any, error)
	Bar(ctx context.Context, i int) (any, error)
}

//@Transactional
func (s *Bar) Foo(ctx context.Context, i any, tx *gorm.DB) (any, error) {
	s.foo.Foo(ctx, i, tx)
	return nil, nil
}

//@Pointcut("validator")
func (s *Bar) Bar(ctx context.Context, i int) (any, error) {
	println(i)
	return i, nil
}
