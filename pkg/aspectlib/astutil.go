package aspectlib

import (
	"bytes"
	"go/ast"
	"go/token"
	"go/types"
)

// File holds a single parsed file and associated data.
type File struct {
	Pkg  *Package  // Package to which this file belongs.
	File *ast.File // Parsed AST.
}

type Package struct {
	Path        string
	Pwd         string
	Name        string
	Defs        map[*ast.Ident]types.Object
	Files       []*File
	OutputFiles map[string][]byte
	FileBuf     map[string]bytes.Buffer
	AspectCache map[string]Aspect
	AspectAlias map[string]Pointcut
	ProxyCache  map[*ast.Ident]Proxy
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
	if ok {
		ident := spec.Name
		if MatchAnnotation(decl.Doc, CommentProxy) {
			p, ok := f.Pkg.ProxyCache[ident]
			if !ok {
				p = &proxy{
					pkg:     f.Pkg,
					name:    ident.String(),
					imports: f.File.Imports,
				}
			}
			params := getCommentParam(decl.Doc, CommentProxy)
			abstract := params[CommentKeyAbstract]
			if v, ok := params[CommentKeyDefault]; ok {
				abstract = v
			}
			suffix := defaultProxySuffix
			if v, ok := params[CommentKeySuffix]; ok {
				suffix = v
			}
			p.SetAbstract(abstract)
			p.SetSuffix(suffix)
			if MatchAnnotation(decl.Doc, CommentPointcut) {
				params := getCommentParam(decl.Doc, CommentPointcut)
				for _, v := range params {
					p.SetPointcuts(&pointcut{name: v})
				}
			}

			f.Pkg.ProxyCache[ident] = p
		}
		// aspect cache
		if MatchAnnotation(decl.Doc, CommentAspect) {
			name := ident.String()
			params := getCommentParam(decl.Doc, CommentAspect)
			fullName := f.Pkg.Name + "." + name
			if p, ok := params[CommentKeyDefault]; ok {
				alias := p
				f.Pkg.AspectAlias[alias] = &pointcut{name: fullName}
			}
			a, ok := f.Pkg.AspectCache[fullName]
			if !ok {
				a = &aspect{
					pkg:     f.Pkg,
					name:    name,
					imports: f.File.Imports,
				}
			}
			f.Pkg.AspectCache[fullName] = a
		}
	}
	return false
}

// funcDecl processes one function declaration clause.
func (f *File) funcDecl(decl *ast.FuncDecl, pkg *Package) bool {
	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return false
	}
	if NoFuncAnnotation(decl.Doc) {
		return false
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
	if MatchAnnotation(decl.Doc, CommentPointcut) {
		method := &method{
			name:    decl.Name.Name,
			params:  decl.Type.Params,
			results: decl.Type.Results,
			f:       decl,
		}
		p, ok := f.Pkg.ProxyCache[ident]
		if !ok {
			p = &proxy{
				pkg:     f.Pkg,
				name:    ident.String(),
				imports: f.File.Imports,
			}
		}
		params := getCommentParam(decl.Doc, CommentPointcut)
		for _, v := range params {
			method.SetPointcuts(&pointcut{name: v})
		}
		p.SetMethods(method)
		f.Pkg.ProxyCache[ident] = p
	} else {
		advice := &advice{
			name: decl.Name.Name,
			f:    decl,
		}
		aspectName := ident.String()
		fullName := f.Pkg.Name + "." + aspectName
		a, ok := f.Pkg.AspectCache[fullName]
		if !ok {
			a = &aspect{
				pkg:     f.Pkg,
				name:    aspectName,
				imports: f.File.Imports,
			}
			f.Pkg.AspectCache[fullName] = a
		}
		// before advice
		if MatchAnnotation(decl.Doc, CommentAdviceBefore) {
			a.SetBefore(advice)
		}
		// after advice
		if MatchAnnotation(decl.Doc, CommentAdviceAfter) {
			a.SetAfter(advice)
		}
		// around advice
		if MatchAnnotation(decl.Doc, CommentAdviceAround) {
			a.SetAround(advice)
		}
	}
	return false
}
