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
type Service struct{}
type IService interface {
	Foo(ctx context.Context, i int) (interface{}, error)
	Bar(ctx context.Context) (string, error)
	Baz(ctx context.Context) (string, error)
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

//@Transactional
func (s Service) Baz(ctx context.Context) (string, error) {
	println("bar")
	return "", nil
}
```

generated code:

```go
type ServiceProxy struct {
	parent IService
}

func NewServiceProxy(parent IService) IService {
	return &ServiceProxy{parent: parent}
}

func (p *ServiceProxy) Foo(ctx context.Context, i int) (r0 interface{}, r1 error) {
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
	return r0, r1
}

func (p *ServiceProxy) Bar(ctx context.Context) (r0 string, r1 error) {
	r0, r1 = p.parent.Bar(ctx)
	return r0, r1
}

func (p *ServiceProxy) Baz(ctx context.Context) (r0 string, r1 error) {
	println("around before trans")
	err := lib.GetGormDB().Transaction(func(tx *gorm.DB) error {
		println("before trans")
		logrus.WithContext(ctx).WithField("func", "Baz").WithField("args", []interface{}{ctx})
		r0, r1 = p.parent.Baz(ctx)
		return r1
	})
	r1 = err
	println("around after trans")
	println("after trans")
	return r0, r1
}
```

aspect definition:

```go
//@Aspect("trans", custom="Transactional")
type AspectTrans struct{}

//@Before
func (a *AspectTrans) Before(jp aspectlib.Joinpoint) {
	println("before trans")
	logrus.WithContext(jp.ParamTo(1).(context.Context)).WithField("func", jp.FuncName()).WithField("args", jp.Params())
}

//@After
func (a *AspectTrans) After(jp aspectlib.Joinpoint) {
	println("after trans")
}

//@Around
func (a *AspectTrans) Around(pjp aspectlib.ProceedingJoinpoint) (result []any) {
	println("around before trans")
	err := lib.GetGormDB().Transaction(func(tx *gorm.DB) error {
		result = pjp.Proceed(pjp.ParamTo(1).(context.Context))
		return pjp.ResultTo(2).(error)
	})
	result[2] = err
	println("around after trans")
	return result
}
```

## todo list

- [x] before advice
- [x] after advice
- [x] around advice
- [x] parameter placeholder
- [x] custom aspect annotation
- [x] factory method for interface
- [ ] dependency injection
