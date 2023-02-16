//go:build sandwich
// +build sandwich

package aspect

import (
	"errors"

	"github.com/go-park/sandwich/pkg/aspect"
)

//@Aspect("validator")
type AspectValidator struct{}

// for func(context.Context, int)(any, error)
//@Around
func (a *AspectValidator) Around(pjp aspect.ProceedingJoinpoint) (any, error) {
	if pjp.ParamTo(2).(int) > 2 {
		r := pjp.ResultTo(1).(any)
		err := errors.New("param i invalid")
		return r, err
	}
	result := pjp.Proceed(pjp.Params()...)
	return result[0], result[1].(error)
}
