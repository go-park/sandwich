package astutils

import (
	"bytes"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/go-park/sandwich/pkg/aspect"
	"github.com/go-park/sandwich/pkg/tools/collections"
	"golang.org/x/tools/go/packages"
)

// File holds a single parsed file and associated data.
type File struct {
	Pkg     *Package  // Package to which this file belongs.
	File    *ast.File // Parsed AST.
	Imports map[string]string
}

type Package struct {
	Path              string
	Pwd               string
	Name              string
	AstPkg            *packages.Package
	Files             []*File
	Defs              map[*ast.Ident]types.Object
	OutputFiles       map[string][]byte
	FileBuf           map[string]bytes.Buffer
	AspectCache       map[string]aspect.Aspect
	AspectAlias       map[string]string
	AspectCustoms     map[Annotation]string
	ProxyCache        map[*ast.Ident]aspect.Proxy
	DelayAspectLoader map[Annotation][]func()
	ComponentCache    map[string]aspect.Component
}

func (p *Package) ImportPath() string {
	if strings.HasSuffix(p.Path, p.Name) {
		return p.Path
	}
	items := strings.Split(p.Path, "/")
	items[len(items)-1] = p.Name
	return strings.Join(items, "/")
}

// InspectGenDecl processes one node.
func (f *File) InspectGenDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok {
		return true
	}
	return f.genDecl(decl, f.Pkg)
}

func (f *File) parseField(fi *ast.Field) (list []aspect.Field) {
	fieldAllPosAnno := parseAnnotation(fi.Doc)
	if !collections.Contains(fieldAllPosAnno, CommentInject) {
		return
	}
	tPkg, tName := getPkgAndName(fi.Type)
	if len(tName) == 0 {
		return
	}
	fullPkg := f.Imports[tPkg]
	// current package
	if len(fullPkg) == 0 {
		fullPkg = f.Pkg.ImportPath()
		tPkg = f.Pkg.Name
	}
	for _, name := range fi.Names {
		f := aspect.NewField(
			aspect.WithFieldName(name.Name),
			aspect.WithFieldType(tPkg, tName),
			aspect.WithFieldInject(fullPkg+"."+tName),
		)
		list = append(list, f)

	}
	return
}

func (f *File) InspectFuncDecl(node ast.Node) bool {
	decl, ok := node.(*ast.FuncDecl)
	if !ok {
		return true
	}
	return f.funcDecl(decl, f.Pkg)
}

// genDecl processes one type declaration clause.
func (f *File) genDecl(decl *ast.GenDecl, pkg *Package) bool {
	if decl.Tok == token.IMPORT {
		for _, v := range decl.Specs {
			if imp, ok := v.(*ast.ImportSpec); ok {
				if len(imp.Path.Value) == 0 {
					return false
				}
				path := strings.ReplaceAll(imp.Path.Value, "\"", "")
				pathList := strings.Split(path, "/")
				name := pathList[len(pathList)-1]
				if imp.Name != nil {
					name = imp.Name.Name
				} else {
					pList, _ := packages.Load(&packages.Config{Mode: packages.NeedName}, path)
					if len(pList) > 0 {
						name = pList[0].Name
					}
				}
				f.Imports[name] = path
			}
		}
		return false
	}
	if !(decl.Tok == token.TYPE) {
		return true
	}
	spec, ok := decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return true
	}
	ident := spec.Name
	allPosAnno := parseAnnotation(decl.Doc)
	if collections.Contains(allPosAnno, CommentProxy) {
		t, ok := spec.Type.(*ast.StructType)
		if !ok {
			return false
		}
		p, ok := f.Pkg.ProxyCache[ident]
		if !ok {
			p = aspect.NewProxy(
				aspect.WithProxyPkg(f.Pkg.Path, f.Pkg.Name),
				aspect.WithProxyName(ident.String()),
				aspect.WithProxyImports(f.File.Imports))
		}
		if t.Fields != nil {
			for _, v := range t.Fields.List {
				fields := f.parseField(v)
				for _, v := range fields {
					p.AddFields(v)
				}
			}
		}
		params := getCommentParam(decl.Doc, CommentProxy)
		abstract := params[CommentKeyAbstract]
		if v, ok := params[CommentKeyDefault]; ok {
			abstract = v
		}
		suffix := DefaultProxySuffix
		if v, ok := params[CommentKeySuffix]; ok {
			suffix = v
		}
		p.SetAbstract(abstract)
		p.SetSuffix(suffix)
		if collections.Contains(allPosAnno, CommentPointcut) {
			params := getCommentParam(decl.Doc, CommentPointcut)
			for _, v := range params {
				p.SetPointcuts(aspect.NewPointcut(aspect.WithPointcutName(v)))
			}
		}
		f.Pkg.ProxyCache[ident] = p
	}

	// aspect cache
	if collections.Contains(allPosAnno, CommentAspect) {
		name := ident.String()
		params := getCommentParam(decl.Doc, CommentAspect)
		fullName := f.Pkg.Name + "." + name
		if alias, ok := params[CommentKeyDefault]; ok {
			f.Pkg.AspectAlias[alias] = fullName
		}
		if custom, ok := params[CommentKeyCustom]; ok {
			if anno, ok := validCustomAnnotation(custom); ok {
				f.Pkg.AspectCustoms[anno] = fullName
			}
		}
		a, ok := f.Pkg.AspectCache[fullName]
		if !ok {
			a = aspect.NewAspect(
				aspect.WithAspectName(name),
				aspect.WithAspectImports(f.File.Imports),
			)
		}
		f.Pkg.AspectCache[fullName] = a
	}
	return false
}

// funcDecl processes one function declaration clause.
func (f *File) funcDecl(decl *ast.FuncDecl, pkg *Package) bool {
	allPosAnno := parseAnnotation(decl.Doc)
	if len(allPosAnno) == 0 {
		return false
	}
	if collections.Contains(allPosAnno, CommentComponent) {
		return f.componentDecl(decl, pkg)
	}
	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return false
	}
	matchCustomAnno := collections.ContainsAny(allPosAnno, collections.Keys(f.Pkg.AspectCustoms)...)
	for _, v := range allPosAnno {
		// not system annotation
		if !IsSystemAnnotation(v) && !matchCustomAnno {
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
	if collections.Contains(allPosAnno, CommentPointcut) || matchCustomAnno {
		method := aspect.NewMethod(aspect.WithMethodDecl(decl))
		p, ok := f.Pkg.ProxyCache[ident]
		if !ok {
			p = aspect.NewProxy(
				aspect.WithProxyPkg(f.Pkg.Path, f.Pkg.Name),
				aspect.WithProxyName(ident.String()),
				aspect.WithProxyImports(f.File.Imports))
		}
		params := getCommentParam(decl.Doc, CommentPointcut)
		for _, v := range params {
			for _, v := range strings.Split(v, ",") {
				method.SetPointcuts(aspect.NewPointcut(aspect.WithPointcutName(v)))
			}
		}
		for _, v := range parseAnnotation(decl.Doc) {
			method.SetPointcuts(aspect.NewPointcut(aspect.WithPointcutName(v.String())))
		}
		p.SetMethods(method)
		f.Pkg.ProxyCache[ident] = p
	}
	// Advice
	if collections.ContainsAny(allPosAnno, AdviceAnnotationList()...) {
		advice := aspect.NewAdvice(aspect.WithAdviceDecl(decl))
		aspectName := ident.String()
		fullName := f.Pkg.Name + "." + aspectName
		a, ok := f.Pkg.AspectCache[fullName]
		if !ok {
			a = aspect.NewAspect(
				aspect.WithAspectName(aspectName),
				aspect.WithAspectImports(f.File.Imports),
			)
			f.Pkg.AspectCache[fullName] = a
		}
		// before advice
		if collections.Contains(allPosAnno, CommentAdviceBefore) {
			a.SetBefore(advice)
		}
		// after advice
		if collections.Contains(allPosAnno, CommentAdviceAfter) {
			a.SetAfter(advice)
		}
		// around advice
		if collections.Contains(allPosAnno, CommentAdviceAround) {
			a.SetAround(advice)
		}
	}
	return false
}

// componentDecl
func (f *File) componentDecl(decl *ast.FuncDecl, pkg *Package) bool {
	results := decl.Type.Results
	if results == nil || len(results.List) != 1 {
		return false
	}
	result := results.List[0]

	compPkgName, compName := getPkgAndName(result.Type)
	if len(compName) == 0 {
		return false
	}
	compPkg := f.Imports[compPkgName]
	// current package
	if len(compPkg) == 0 {
		compPkg = pkg.ImportPath()
		compPkgName = pkg.Name
	}
	comp := aspect.NewComponent(
		aspect.WithComponentFactory(pkg.Path, decl.Name.Name),
		aspect.WithComponentPkg(compPkg, compPkgName),
		aspect.WithComponentName(compPkg+"."+compName),
	)
	f.Pkg.ComponentCache[comp.Name()] = comp
	return true
}
