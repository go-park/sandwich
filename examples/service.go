package main

import (
	"context"

	"github.com/go-park/sandwich/examples/lib"
	"gorm.io/gorm"
)

var _ IService = &Service{}

//@Proxy("IService")
type Service struct {
	//@Inject
	foo lib.Foo
}
type IService interface {
	Foo(ctx context.Context, i interface{}, tx *gorm.DB) (interface{}, error)
	Bar(ctx context.Context) (string, error)
	Baz(ctx context.Context, i interface{}, tx *gorm.DB) (string, error)
}

//@Pointcut("log", "trans")
func (s *Service) Foo(ctx context.Context, i interface{}, tx *gorm.DB) (interface{}, error) {
	println("foo")
	return nil, nil
}

//@Pointcut
func (s Service) Bar(ctx context.Context) (string, error) {
	println("bar")
	return "", nil
}

//@Transactional
func (s Service) Baz(ctx context.Context, i interface{}, tx *gorm.DB) (string, error) {
	println("bar")
	return "", nil
}
