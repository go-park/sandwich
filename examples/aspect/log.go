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
func (a *AspectLog) Around(pjp aspect.ProceedingJoinpoint) []interface{} {
	fmt.Println("around before log")
	fmt.Println("params: ", pjp.Params())
	result := pjp.Proceed()
	fmt.Println("results: ", pjp.Results(), pjp.ResultTo(2).(error))
	fmt.Println("around after log")
	return result
}
