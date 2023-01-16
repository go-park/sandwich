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

*foo.go*
```go
var _ IFoo = &Foo{}

//@Proxy("IFoo")
type Foo struct {
	//@Inject
	foo lib.Foo
}
type IFoo interface {
	Foo(ctx context.Context, i any, tx *gorm.DB) (any, error)
}

//@Pointcut("log", "trans")
func (s *Foo) Foo(ctx context.Context, i any, tx *gorm.DB) (any, error) {
	println("foo")
	return nil, nil
}
```

*bar.go*
```go
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
}

//@Transactional
func (s *Bar) Foo(ctx context.Context, i any, tx *gorm.DB) (any, error) {
	s.foo.Foo(ctx, i, tx)
	return nil, nil
}
```

generated code:

*foo_proxy.gen.go*
```go
type FooProxy struct {
	parent Foo
}

func NewFooProxy() IFoo {
	return &Foo{
		foo: lib.NewFoo(),
	}
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
```

*bar_proxy.gen.go*
```go
type BarProxy struct {
	parent Bar
}

func NewBarProxy() IBar {
	return &Bar{
		foo:    NewFooProxy(),
		libFoo: lib.NewFoo(),
	}
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
```

## todo list

- [x] before advice
- [x] after advice
- [x] around advice
- [x] parameter placeholder
- [x] custom aspect annotation
- [x] factory method for interface
- [x] dependency injection
