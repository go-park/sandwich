package aspect

import (
	"go/ast"
	"strings"
)

type (
	Option[T any] func(*T)
)

func WithProxyPkg(path, name string) Option[proxy] {
	return func(o *proxy) {
		o.pkgPath = path
		o.pkgName = name
	}
}

func WithProxyName(name string) Option[proxy] {
	return func(o *proxy) {
		o.name = name
	}
}

func WithProxyImports(specs []*ast.ImportSpec) Option[proxy] {
	return func(o *proxy) {
		o.imports = specs
	}
}

func WithAspectName(name string) Option[aspect] {
	return func(o *aspect) {
		o.name = name
	}
}

func WithAspectImports(specs []*ast.ImportSpec) Option[aspect] {
	return func(o *aspect) {
		o.imports = specs
	}
}

func WithPointcutName(name string) Option[pointcut] {
	return func(o *pointcut) {
		o.name = name
	}
}

func WithAdviceName(name string) Option[advice] {
	return func(o *advice) {
		o.name = name
	}
}

func WithAdviceDecl(decl *ast.FuncDecl) Option[advice] {
	return func(o *advice) {
		o.f = decl
		o.name = decl.Name.Name
	}
}

func WithMethodName(name string) Option[method] {
	return func(o *method) {
		o.name = name
	}
}

func WithMethodParams(params *ast.FieldList) Option[method] {
	return func(o *method) {
		o.params = params
	}
}

func WithMethodResults(results *ast.FieldList) Option[method] {
	return func(o *method) {
		o.results = results
	}
}

func WithMethodDecl(decl *ast.FuncDecl) Option[method] {
	return func(o *method) {
		o.f = decl
		o.name = decl.Name.Name
		o.params = decl.Type.Params
		o.results = decl.Type.Results
	}
}

func WithComponentPkg(path, name string) Option[component] {
	return func(o *component) {
		o.pkgPath = path
		o.pkgName = name
	}
}

func WithComponentName(name string) Option[component] {
	return func(o *component) {
		o.name = name
	}
}

func WithComponentFactory(pkg, name string) Option[component] {
	return func(c *component) {
		c.factoryPkg = pkg
		c.factoryName = name
	}
}

func WithFieldName(name string) Option[field] {
	return func(c *field) {
		c.name = name
	}
}

func WithFieldType(pkg, name string) Option[field] {
	return func(c *field) {
		if strings.HasPrefix(name, "*") {
			name = strings.TrimPrefix(name, "*")
			pkg = "*" + pkg
		}
		c.typ = name
		c.tPkg = pkg
	}
}

func WithFieldInject(name string) Option[field] {
	return func(c *field) {
		c.inject = name
	}
}
