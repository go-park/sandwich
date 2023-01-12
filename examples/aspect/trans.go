package aspect

import (
	"context"

	"github.com/go-park/sandwich/examples/lib"
	"github.com/go-park/sandwich/pkg/aspect"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//@Aspect("trans", custom="Transactional")
type AspectTrans struct{}

//@Before
func (a *AspectTrans) Before(jp aspect.Joinpoint) {
	println("before trans")
	logrus.WithContext(jp.ParamTo(1).(context.Context)).WithField("func", jp.FuncName()).WithField("args", jp.Params())
}

//@After
func (a *AspectTrans) After(jp aspect.Joinpoint) {
	println("after trans")
}

//@Around
func (a *AspectTrans) Around(pjp aspect.ProceedingJoinpoint) (result []any) {
	println("around before trans")
	err := lib.GetGormDB().Transaction(func(tx1 *gorm.DB) error {
		result = pjp.Proceed()
		return pjp.ResultTo(2).(error)
	})
	result[2] = err
	println("around after trans")
	return result
}
