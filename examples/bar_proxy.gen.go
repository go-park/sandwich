// Code generated by sandwich. DO NOT EDIT.

package main

import (
	"context"
	"errors"

	"github.com/go-park/sandwich/examples/lib"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BarProxy struct {
	parent Bar
}

//@Component
func NewBarProxy() IBar {
	pa := &Bar{
		foo:    NewFooProxy(),
		libFoo: lib.NewFoo(),
	}

	return pa
}

func (p *BarProxy) Foo(ctx context.Context, i any, tx *gorm.DB) (r0 any, r1 error) {
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
	return r0, r1
}

func (p *BarProxy) Bar(ctx context.Context, i int) (r0 any, r1 error) {
	if i > 2 {
		r := r0
		err := errors.New("param i invalid")
		return r, err
	}
	r0, r1 = p.parent.Bar(ctx, i)
	return r0, r1
}
