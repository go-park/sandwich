package aspect

import (
	"fmt"

	"github.com/go-park/sandwich/pkg/aspectlib"
)

//@Aspect("log")
type AspectLog struct{}

//@Before
func (a *AspectLog) Before(jp aspectlib.Joinpoint) {
	fmt.Println("before log")
}

//@After
func (a *AspectLog) After(jp aspectlib.Joinpoint) {
	fmt.Println("after log")
}

//@Around
func (a *AspectLog) Around(pjp aspectlib.ProceedingJoinpoint) []interface{} {
	fmt.Println("around before log")
	fmt.Println("params: ", pjp.Params())
	result := pjp.Proceed()
	fmt.Println("results: ", pjp.Results(), pjp.ResultTo(2).(error))
	fmt.Println("around after log")
	return result
}
