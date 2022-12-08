package astutils

import (
	"bytes"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/go-park/sandwich/pkg/aspectlib"
	"github.com/go-park/sandwich/pkg/tools.go/collections"
)

// File holds a single parsed file and associated data.
type File struct {
	Pkg  *Package  // Package to which this file belongs.
	File *ast.File // Parsed AST.
}

type Package struct {
	Path              string
	Pwd               string
	Name              string
	Defs              map[*ast.Ident]types.Object
	Files             []*File
	OutputFiles       map[string][]byte
	FileBuf           map[string]bytes.Buffer
	AspectCache       map[string]aspectlib.Aspect
	AspectAlias       map[string]string
	AspectCustoms     map[aspectlib.Annotation]string
	ProxyCache        map[*ast.Ident]aspectlib.Proxy
	DelayAspectLoader map[aspectlib.Annotation][]func()
}

// inspectDecl processes one node.
func (f *File) InspectDecl(node ast.Node) bool {
	gdecl, gOk := node.(*ast.GenDecl)
	fdecl, fOk := node.(*ast.FuncDecl)
	if !gOk && !fOk {
		return true
	}
	if gOk {
		return f.genDecl(gdecl, f.Pkg)
	}
	if fOk {
		return f.funcDecl(fdecl, f.Pkg)
	}
	return false
}

// genDecl processes one type declaration clause.
func (f *File) genDecl(decl *ast.GenDecl, pkg *Package) bool {
	if !(decl.Tok == token.TYPE) {
		return true
	}
	spec, ok := decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return true
	}
	ident := spec.Name
	allPosAnno := parseAnnotation(decl.Doc)
	if collections.Contains(allPosAnno, aspectlib.CommentProxy) {
		p, ok := f.Pkg.ProxyCache[ident]
		if !ok {
			p = aspectlib.NewProxy(
				aspectlib.WithProxyPkg(f.Pkg.Path, f.Pkg.Name),
				aspectlib.WithProxyName(ident.String()),
				aspectlib.WithProxyImports(f.File.Imports))
		}
		params := getCommentParam(decl.Doc, aspectlib.CommentProxy)
		abstract := params[aspectlib.CommentKeyAbstract]
		if v, ok := params[aspectlib.CommentKeyDefault]; ok {
			abstract = v
		}
		suffix := aspectlib.DefaultProxySuffix
		if v, ok := params[aspectlib.CommentKeySuffix]; ok {
			suffix = v
		}
		p.SetAbstract(abstract)
		p.SetSuffix(suffix)
		if collections.Contains(allPosAnno, aspectlib.CommentPointcut) {
			params := getCommentParam(decl.Doc, aspectlib.CommentPointcut)
			for _, v := range params {
				p.SetPointcuts(aspectlib.NewPointcut(aspectlib.WithPointcutName(v)))
			}
		}
		f.Pkg.ProxyCache[ident] = p
	}

	// aspect cache
	if collections.Contains(allPosAnno, aspectlib.CommentAspect) {
		name := ident.String()
		params := getCommentParam(decl.Doc, aspectlib.CommentAspect)
		fullName := f.Pkg.Name + "." + name
		if alias, ok := params[aspectlib.CommentKeyDefault]; ok {
			f.Pkg.AspectAlias[alias] = fullName
		}
		if custom, ok := params[aspectlib.CommentKeyCustom]; ok {
			if anno, ok := validCustomAnnotation(custom); ok {
				f.Pkg.AspectCustoms[anno] = fullName
			}
		}
		a, ok := f.Pkg.AspectCache[fullName]
		if !ok {
			a = aspectlib.NewAspect(
				aspectlib.WithAspectName(name),
				aspectlib.WithAspectImports(f.File.Imports),
			)
		}
		f.Pkg.AspectCache[fullName] = a
	}
	return false
}

// funcDecl processes one function declaration clause.
func (f *File) funcDecl(decl *ast.FuncDecl, pkg *Package) bool {
	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return false
	}
	allPosAnno := parseAnnotation(decl.Doc)
	if len(allPosAnno) == 0 {
		return false
	}
	matchCustomAnno := collections.ContainsAny(allPosAnno, collections.Keys(f.Pkg.AspectCustoms)...)
	for _, v := range allPosAnno {
		// not system annotation
		if !aspectlib.IsSystemAnnotation(v) && !matchCustomAnno {
			f.Pkg.DelayAspectLoader[v] = append(f.Pkg.DelayAspectLoader[v], func() { f.funcDecl(decl, pkg) })
		}
	}
	recv := decl.Recv.List[0]
	var ident *ast.Ident
	var aspectSpec *ast.TypeSpec
	if star, ok := recv.Type.(*ast.StarExpr); ok {
		aspectSpec = star.X.(*ast.Ident).Obj.Decl.(*ast.TypeSpec)
	} else if id, ok := recv.Type.(*ast.Ident); ok {
		aspectSpec = id.Obj.Decl.(*ast.TypeSpec)
	}
	ident = aspectSpec.Name
	// Pointcut
	if collections.Contains(allPosAnno, aspectlib.CommentPointcut) || matchCustomAnno {
		method := aspectlib.NewMethod(aspectlib.WithMethodDecl(decl))
		p, ok := f.Pkg.ProxyCache[ident]
		if !ok {
			p = aspectlib.NewProxy(
				aspectlib.WithProxyPkg(f.Pkg.Path, f.Pkg.Name),
				aspectlib.WithProxyName(ident.String()),
				aspectlib.WithProxyImports(f.File.Imports))
		}
		params := getCommentParam(decl.Doc, aspectlib.CommentPointcut)
		for _, v := range params {
			method.SetPointcuts(aspectlib.NewPointcut(aspectlib.WithPointcutName(v)))
		}
		for _, v := range parseAnnotation(decl.Doc) {
			method.SetPointcuts(aspectlib.NewPointcut(aspectlib.WithPointcutName(v.String())))
		}
		p.SetMethods(method)
		f.Pkg.ProxyCache[ident] = p
	}
	// Advice
	if collections.ContainsAny(allPosAnno, aspectlib.AdviceAnnotationList()...) {
		advice := aspectlib.NewAdvice(aspectlib.WithAdviceDecl(decl))
		aspectName := ident.String()
		fullName := f.Pkg.Name + "." + aspectName
		a, ok := f.Pkg.AspectCache[fullName]
		if !ok {
			a = aspectlib.NewAspect(
				aspectlib.WithAspectName(aspectName),
				aspectlib.WithAspectImports(f.File.Imports),
			)
			f.Pkg.AspectCache[fullName] = a
		}
		// before advice
		if collections.Contains(allPosAnno, aspectlib.CommentAdviceBefore) {
			a.SetBefore(advice)
		}
		// after advice
		if collections.Contains(allPosAnno, aspectlib.CommentAdviceAfter) {
			a.SetAfter(advice)
		}
		// around advice
		if collections.Contains(allPosAnno, aspectlib.CommentAdviceAround) {
			a.SetAround(advice)
		}
	}
	return false
}
