# sandwich

go aop and ioc tools based on code generation

## Usage

### Install

```shell
# golang 1.18+ required
go install github.com/go-park/sandwich/cmd/aspect@latest
```

### Annotation

`@Proxy` for struct generate a file with _gen.go suffix

`@Aspect` for struct use to enhance the proxy struct

`@Before` for struct function use to enhance the proxy struct function

`@After` for struct function use to enhance the proxy struct function

`@Around` for struct function use to enhance the proxy struct function

`@Component` for struct factory method use to inject the proxy struct

`@Pointcut` for struct function generate a proxy func for proxy struct

`@Inject` for struct field use to inject proxy struct

### Usage

```shell
git clone https://github.com/go-park/sandwich.git
cd sandwich/examples
# add "go:generate aspect -tags=sandwich ." comment to the main function for go generate
go run ./... -tags=sandwich . # aspect .
```

aspect example:

```go
//go:build sandwich
// +build sandwich

package aspect

import (
	"fmt"

	"github.com/go-park/sandwich/pkg/aspect"
)

//@Aspect("log")
type AspectLog struct{}

//@Before
func (a *AspectLog) Before(jp aspect.Joinpoint) {
	fmt.Println("before log")
}

//@After
func (a *AspectLog) After(jp aspect.Joinpoint) {
	fmt.Println("after log")
}

//@Around
func (a *AspectLog) Around(pjp aspect.ProceedingJoinpoint) []any {
	fmt.Println("around before log")
	fmt.Println("params: ", pjp.Params())
	result := pjp.Proceed()
	fmt.Println("results: ", pjp.Results(), pjp.ResultTo(2).(error))
	fmt.Println("around after log")
	return result
}
```

raw code:

*foo.go*
```go
var _ IFoo = &Foo{}

//@Proxy("IFoo")
type Foo struct {
	//@Inject
	foo lib.Foo
	//@Value("123")
	str string
	//@Value("true")
	boo bool
	//@Value("123")
	num uint64
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
```

generate code:

*foo_proxy.gen.go*
```go
type FooProxy struct {
	parent Foo
}

//@Component
func NewFooProxy() IFoo {
	return &Foo{
		foo: lib.NewFoo(),
		str: "123",
		boo: true,
		num: 123,
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
```

## todo list

- [x] before advice
- [x] after advice
- [x] around advice
- [x] parameter placeholder
- [x] custom aspect annotation
- [x] factory method for interface
- [x] dependency injection
- [x] proxy interception
