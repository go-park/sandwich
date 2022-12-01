package examples

import (
	"context"
)

//@Proxy
type Service struct{}

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
