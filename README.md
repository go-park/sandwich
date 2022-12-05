# sandwich

aop library based on code generation for golang

## Usage

insatll:

```shell
# golang 1.18+ required
go install github.com/go-park/sandwich/cmd/aspect@latest
```

examples:

```shell
git clone https://github.com/go-park/sandwich.git
cd sandwich/examples
go run ../cmd/aspect . # aspect .
```

raw code:

```go
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
```

generated code:

```go
type ServiceProxy struct {
	parent *Service
}

func NewServiceProxy(parent *Service) *ServiceProxy {
	return &ServiceProxy{parent: parent}
}

func (p *ServiceProxy) Foo(ctx context.Context, i int) (r0 interface{}, r1 error) {
	fmt.Println("around before log")
	fmt.Println("params: ", []interface{}{ctx, i})
	fmt.Println("before log")
	println("around before trans")
	err := lib.GetGormDB().Transaction(func(tx *gorm.DB) error {
		println("before trans")
		logrus.WithContext(ctx).WithField("func", "Foo").WithField("args", []interface{}{ctx, i})
		r0, r1 = p.parent.Foo(ctx, i)
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

func (p *ServiceProxy) Bar(ctx context.Context) (r0 string, r1 error) {
	r0, r1 = p.parent.Bar(ctx)
	return r0, r1
}

```

## todo list

- [x] before advice
- [x] after advice
- [x] parameter placeholder
- [x] around advice
- [ ] custom annotation
- [X] factory method for interface
