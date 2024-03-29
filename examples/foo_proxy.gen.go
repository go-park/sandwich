// Code generated by sandwich. DO NOT EDIT.

package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-park/sandwich/examples/lib"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type FooProxy struct {
	parent Foo
}

var (
	_IFooInst IFoo
	_IFooOnce sync.Once
)

//@Component
func NewFooProxy() IFoo {
	_IFooOnce.Do(func() {
		_IFooInst = &Foo{
			foo: lib.NewFoo(),
			str: "123",
			boo: true,
			num: 123,
		}
	})
	return _IFooInst
}

func (p *FooProxy) Foo(ctx context.Context, i any, tx *gorm.DB) (r0 any, r1 error) {
	fmt.Println("around before log")
	fmt.Println("params: ", []interface{}{ctx, i, tx})
	fmt.Println("before log")
	println("around before trans")
	err := lib.GetGormDB().Transaction(func(tx1 *gorm.DB) error {
		println("before trans")
		logrus.WithContext(ctx).WithField("func", "Foo").WithField("args", []interface{}{ctx, i, tx})
		r0, r1 = p.parent.Foo(ctx, i, tx)
		return r1
	})
	r1 = err
	println("around after trans")
	println("after trans")
	fmt.Println("results: ", []interface{}{r0, r1}, r1)
	fmt.Println("around after log")
	fmt.Println("after log")
	return r0, r1
}
