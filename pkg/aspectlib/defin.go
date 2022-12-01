package aspectlib

import (
	"fmt"
	"go/ast"
	"strings"
)

var (
	_ Proxy  = (*proxy)(nil)
	_ Method = (*method)(nil)
)

type (
	Nameable interface {
		Name() string
	}

	Cutable interface {
		SetPointcuts(po ...Pointcut)
		GetPointcuts() []Pointcut
	}

	// Proxy
	Proxy interface {
		Nameable
		Cutable
		SetMethods(m ...Method)
		GetMethods() []Method
		Pkg() *Package
		Imports() []*ast.ImportSpec
	}

	// Pointcut
	Pointcut interface {
		Nameable
	}

	// Advice
	Advice interface {
		Nameable
		Func() *ast.FuncDecl
	}

	// Aspect
	Aspect interface {
		Nameable
		SetBefore(Advice)
		SetAfter(Advice)
		SetAround(Advice)
		GetBefore() Advice
		GetAfter() Advice
		GetAround() Advice
		Pkg() *Package
		Imports() []*ast.ImportSpec
	}

	// Method
	Method interface {
		Nameable
		Cutable
		GetParams() ([]string, []string)
		GetResults() ([]string, []string)
	}

	// Joinpoint
	Joinpoint interface {
		Nameable
		ParamTo(i int) any
		Params() []any
		Results() []any
		ResultTo(i int) any
		FuncName() string
	}

	// ProceedingJoinpoint
	ProceedingJoinpoint interface {
		Joinpoint
		Proceed(...any) []any
		ProceedOneResult(...any) []any
	}
)

type (
	// implement Proxy
	proxy struct {
		pkg       *Package
		name      string
		methods   []Method
		pointcuts []Pointcut
		imports   []*ast.ImportSpec
	}
	// implement Pointcut
	pointcut struct {
		name string
	}
	// implement Advice
	advice struct {
		name string
		f    *ast.FuncDecl
	}
	// implement Aspect
	aspect struct {
		pkg     *Package
		name    string
		before  Advice
		after   Advice
		around  Advice
		imports []*ast.ImportSpec
	}
	// implement Method
	method struct {
		name      string
		f         *ast.FuncDecl
		params    *ast.FieldList
		results   *ast.FieldList
		pointcuts []Pointcut
	}
)

func (p *proxy) Name() string    { return p.name }
func (p *aspect) Name() string   { return p.name }
func (p *method) Name() string   { return p.name }
func (p *pointcut) Name() string { return p.name }
func (p *advice) Name() string   { return p.name }

func (p *proxy) Pkg() *Package  { return p.pkg }
func (p *aspect) Pkg() *Package { return p.pkg }

func (p *proxy) Imports() []*ast.ImportSpec  { return p.imports }
func (p *aspect) Imports() []*ast.ImportSpec { return p.imports }

func (p *proxy) SetMethods(m ...Method) {
	p.methods = append(p.methods, m...)
}

func (p *proxy) SetPointcuts(po ...Pointcut) {
	p.pointcuts = append(p.pointcuts, po...)
}

func (p *proxy) GetMethods() []Method {
	return p.methods
}

func (p *proxy) GetPointcuts() []Pointcut {
	return p.pointcuts
}

func (p *aspect) GetBefore() Advice {
	return p.before
}

func (p *aspect) GetAfter() Advice {
	return p.after
}

func (p *aspect) GetAround() Advice {
	return p.around
}

func (p *aspect) SetBefore(before Advice) {
	p.before = before
}

func (p *aspect) SetAfter(after Advice) {
	p.after = after
}

func (p *aspect) SetAround(around Advice) {
	p.around = around
}

func (p *advice) Func() *ast.FuncDecl { return p.f }

func (p *method) GetParams() ([]string, []string) {
	return p.parseFields(p.params)
}

func (p *method) GetResults() ([]string, []string) {
	return p.parseFields(p.results)
}

func (p *method) parseFields(paramOrResult *ast.FieldList) ([]string, []string) {
	var paramNames, params []string
	if paramOrResult == nil {
		return paramNames, params
	}
	for i, param := range paramOrResult.List {
		var names []string
		for _, v := range param.Names {
			names = append(names, v.Name)
		}
		if len(names) == 0 {
			names = append(names, fmt.Sprintf("r%d", i))
		}
		paramNames = append(paramNames, names...)
		var paramType string
		if ident, ok := param.Type.(*ast.Ident); ok {
			paramType = ident.Name
		}
		if inter, ok := param.Type.(*ast.InterfaceType); ok {
			paramType = "interface{%s}"
			var methods []string
			// todo fill interface methods
			for _, v := range inter.Methods.List {
				_ = v
			}
			paramType = fmt.Sprintf(paramType, strings.Join(methods, "\n"))
		}
		if sel, ok := param.Type.(*ast.SelectorExpr); ok {
			paramType = fmt.Sprintf("%s.%s", sel.X.(*ast.Ident).Name, sel.Sel.Name)
		}
		pa := fmt.Sprintf("%s %s",
			strings.Join(names, ","), paramType,
		)
		params = append(params, strings.TrimSpace(pa))
	}
	return paramNames, params
}

func (p *method) SetPointcuts(po ...Pointcut) {
	p.pointcuts = append(p.pointcuts, po...)
}

func (p *method) GetPointcuts() []Pointcut {
	return p.pointcuts
}
