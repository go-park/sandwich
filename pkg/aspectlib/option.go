package aspectlib

import "go/ast"

type (
	ProxyOption    func(*proxy)
	AspectOption   func(*aspect)
	PointcutOption func(*pointcut)
	AdviceOption   func(*advice)
	MethodOption   func(*method)
)

func WithProxyPkg(path, name string) ProxyOption {
	return func(o *proxy) {
		o.pkgPath = path
		o.pkgName = name
	}
}

func WithProxyName(name string) ProxyOption {
	return func(o *proxy) {
		o.name = name
	}
}

func WithProxyImports(specs []*ast.ImportSpec) ProxyOption {
	return func(o *proxy) {
		o.imports = specs
	}
}

func WithAspectName(name string) AspectOption {
	return func(o *aspect) {
		o.name = name
	}
}

func WithAspectImports(specs []*ast.ImportSpec) AspectOption {
	return func(o *aspect) {
		o.imports = specs
	}
}

func WithPointcutName(name string) PointcutOption {
	return func(o *pointcut) {
		o.name = name
	}
}

func WithAdviceName(name string) AdviceOption {
	return func(o *advice) {
		o.name = name
	}
}

func WithAdviceDecl(decl *ast.FuncDecl) AdviceOption {
	return func(o *advice) {
		o.f = decl
		o.name = decl.Name.Name
	}
}

func WithMethodName(name string) MethodOption {
	return func(o *method) {
		o.name = name
	}
}

func WithMethodParams(params *ast.FieldList) MethodOption {
	return func(o *method) {
		o.params = params
	}
}

func WithMethodResults(results *ast.FieldList) MethodOption {
	return func(o *method) {
		o.results = results
	}
}

func WithMethodDecl(decl *ast.FuncDecl) MethodOption {
	return func(o *method) {
		o.f = decl
		o.name = decl.Name.Name
		o.params = decl.Type.Params
		o.results = decl.Type.Results
	}
}
