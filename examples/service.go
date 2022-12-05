package examples

import (
	"context"
)

var _ IService = &Service{}

//@Proxy("IService")
type Service struct{}
type IService interface {
	Foo(ctx context.Context, i int) (interface{}, error)
	Bar(ctx context.Context) (string, error)
}

func NewService() IService {
	return &Service{}
}

//@Pointcut("log", "trans")
func (s *Service) Foo(ctx context.Context, i int) (interface{}, error) {
	println("foo")
	return nil, nil
}

//@Pointcut
func (s Service) Bar(ctx context.Context) (string, error) {
	println("bar")
	return "", nil
}
