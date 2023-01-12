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
# add "go:generate aspect ." comment to the main function for go generate
go run ../cmd/aspect . # aspect .
```

raw code:

```go
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

```

generated code:

```go
type ServiceProxy struct {
	parent Service
}

func NewServiceProxy() IService {
	return &Service{
		foo: lib.NewFoo(),
	}
}

func (p *ServiceProxy) Foo(ctx context.Context, i interface{}, tx *gorm.DB) (r0 interface{}, r1 error) {
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

func (p *ServiceProxy) Bar(ctx context.Context) (r0 string, r1 error) {
	r0, r1 = p.parent.Bar(ctx)
	return r0, r1
}

func (p *ServiceProxy) Baz(ctx context.Context, i interface{}, tx *gorm.DB) (r0 string, r1 error) {
	println("around before trans")
	err := lib.GetGormDB().Transaction(func(tx1 *gorm.DB) error {
		println("before trans")
		logrus.WithContext(ctx).WithField("func", "Baz").WithField("args", []interface{}{ctx, i, tx})
		r0, r1 = p.parent.Baz(ctx, i, tx)
		return r1
	})
	r1 = err
	println("around after trans")
	println("after trans")
	return r0, r1
}
```

## todo list

- [x] before advice
- [x] after advice
- [x] around advice
- [x] parameter placeholder
- [x] custom aspect annotation
- [x] factory method for interface
- [x] dependency injection
