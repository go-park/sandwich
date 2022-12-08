package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/go-park/sandwich/pkg/aspectlib"
	"github.com/go-park/sandwich/pkg/astutils"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	options
	pkgList           map[string]*astutils.Package // Package we are scanning.
	aspectCache       map[string]aspectlib.Aspect
	aspectAlias       map[string]string
	aspectCustoms     map[aspectlib.Annotation]string
	proxyCache        map[*ast.Ident]aspectlib.Proxy
	delayAspectLoader map[aspectlib.Annotation][]func()
}

func NewGenerator(opts ...Option) *Generator {
	ge := &Generator{
		options:           DefaultOptions(),
		pkgList:           map[string]*astutils.Package{},
		aspectCache:       map[string]aspectlib.Aspect{},
		aspectAlias:       map[string]string{},
		aspectCustoms:     map[aspectlib.Annotation]string{},
		proxyCache:        map[*ast.Ident]aspectlib.Proxy{},
		delayAspectLoader: map[aspectlib.Annotation][]func(){},
	}
	for _, opt := range opts {
		opt.apply(&ge.options)
	}
	return ge
}

// parsePackage recursive analyzes the package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func (g *Generator) ParsePackage() *Generator {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedTypesInfo |
			packages.NeedSyntax |
			packages.NeedDeps |
			packages.NeedImports |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(g.tags, " "))},
		Logf:       log.Printf,
	}
	if g.recursive {
		g.patterns = getAllPathPatterns(g.patterns)
	}
	pkgList, err := packages.Load(cfg, g.patterns...)
	if err != nil {
		log.Fatal(err)
	}
	var depPkgList []*packages.Package
	for _, dep := range g.deps {
		for _, pkg := range pkgList {
			for k, v := range pkg.Imports {
				if strings.HasPrefix(k, dep) {
					depPkgList = append(depPkgList, v)
				}
			}
		}
	}
	log.Printf("package patterns is %v, load result is %v, dep result is %v", g.patterns, pkgList, depPkgList)
	g.addPackage(append(pkgList, depPkgList...)...)
	return g
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(list ...*packages.Package) {
	for _, pkg := range list {
		log.Println(pkg.Imports)
		item := &astutils.Package{
			Name:              pkg.Name,
			Path:              pkg.PkgPath,
			Pwd:               getCurrentPkg(),
			Defs:              map[*ast.Ident]types.Object{},
			Files:             make([]*astutils.File, len(pkg.Syntax)),
			OutputFiles:       map[string][]byte{},
			FileBuf:           map[string]bytes.Buffer{},
			AspectCache:       g.aspectCache,
			AspectAlias:       g.aspectAlias,
			AspectCustoms:     g.aspectCustoms,
			ProxyCache:        g.proxyCache,
			DelayAspectLoader: g.delayAspectLoader,
		}
		for i, file := range pkg.Syntax {
			item.Files[i] = &astutils.File{
				File: file,
				Pkg:  item,
			}
		}
		g.pkgList[item.Path] = item
	}
}

// Generate inspect node and construct proxy data
func (g *Generator) Generate() *Generator {
	for _, pkg := range g.pkgList {
		for _, file := range pkg.Files {
			if file.File != nil {
				ast.Inspect(file.File, file.InspectDecl)
			}
		}
	}
	for anno := range g.aspectCustoms {
		for _, load := range g.delayAspectLoader[anno] {
			load()
		}
	}

	for k, proxy := range g.proxyCache {
		if !k.IsExported() {
			log.Panic("unexported method cannot be proxy")
		}
		abstract := proxy.Abstract()
		parent := abstract
		if len(abstract) == 0 {
			parent = proxy.Name()
			abstract = "*" + proxy.Name() + proxy.Suffix()
		}
		pd := aspectlib.ProxyData{
			Package:         proxy.PkgName(),
			ProxyStructName: proxy.Name() + proxy.Suffix(),
			AbstractName:    abstract,
			ParentName:      parent,
		}
		pd.Imports = append(pd.Imports, astutils.GetImports(proxy.Imports())...)
		cuts := proxy.GetPointcuts()
		for _, method := range proxy.GetMethods() {
			cuts := append(cuts, method.GetPointcuts()...)
			paramNames, params := method.GetParams()
			resultNames, results := method.GetResults()
			m := &aspectlib.ProxyMethod{
				Name:        method.Name(),
				Params:      strings.Join(params, ", "),
				ParamNames:  strings.Join(paramNames, ", "),
				Results:     strings.Join(results, ", "),
				ResultNames: strings.Join(resultNames, ", "),
			}
			var postStack [][]string
			for _, cut := range cuts {
				aspectName := cut.Name()
				if alias, ok := g.aspectAlias[aspectName]; ok {
					aspectName = alias
				} else if anno, ok := g.aspectCustoms[aspectlib.Annotation(aspectName)]; ok {
					aspectName = anno
				}

				aspect, ok := g.aspectCache[aspectName]
				if !ok {
					continue
				}
				pd.Imports = append(pd.Imports, astutils.GetImports(aspect.Imports())...)
				before := astutils.ParseAdviceStmt(aspect.GetBefore(), method)
				after := astutils.ParseAdviceStmt(aspect.GetAfter(), method)
				aroundBefore, aroundAfter := astutils.ParseAroundAdvice(aspect.GetAround(), method)
				postStack = append(postStack, after, aroundAfter)
				for _, v := range append(aroundBefore, before...) {
					if strings.HasPrefix(v, "-") {
						continue
					}
					m.Before = append(m.Before, template.HTML(v))
				}
			}
			// invoke method of proxy
			args, _ := method.GetParams()
			proceedStmt := fmt.Sprintf("p.parent.%s(%s)", method.Name(), strings.Join(args, ", "))
			rets, _ := method.GetResults()
			if len(rets) > 0 {
				proceedStmt = fmt.Sprintf("%s = %s", strings.Join(rets, ", "), proceedStmt)
			}
			postStack = append(postStack, []string{proceedStmt})
			// reverse after/around advice
			for len(postStack) > 0 {
				n := len(postStack) - 1
				for _, v := range postStack[n] {
					if strings.HasPrefix(v, "-") {
						continue
					}
					m.After = append(m.After, template.HTML(v))
				}
				postStack = postStack[:n]
			}
			pd.Methods = append(pd.Methods, m)
		}
		tpl, err := template.New("").Parse(aspectlib.GetProxyTpl())
		if err != nil {
			panic(err.Error())
		}
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, pd); err != nil {
			panic(err.Error())
		}
		g.pkgList[proxy.PkgPath()].FileBuf[proxy.Name()] = buf
	}
	return g
}

// Format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) Format() *Generator {
	for _, pkg := range g.pkgList {
		for k, v := range pkg.FileBuf {
			src, err := imports.Process("", v.Bytes(), nil)
			if err != nil {
				log.Println("output file:\n", string(v.Bytes()))
				// Should never happen, but can arise when developing this code.
				// The user can compile the output to see the error.
				log.Printf("warning: internal error: invalid Go generated: %s", err)
				log.Printf("warning: compile the package to analyze the error")
				continue
			}
			pkg.OutputFiles[k] = src
		}
	}
	return g
}

// Output create proxy file by Generator's buffer.
func (g *Generator) Output() *Generator {
	for _, pkg := range g.pkgList {
		for k, v := range pkg.OutputFiles {
			// Write to file.
			outputName := ""
			if outputName == "" {
				baseName := fmt.Sprintf("%s_proxy.gen.go", k)
				targetPkg := getRelevantPkg(pkg.Pwd, pkg.Path)
				log.Printf("current pkg is %s target pkg is %s target relevant pkg is %s", pkg.Pwd, pkg.Path, targetPkg)
				outputName = filepath.Join(targetPkg, strings.ToLower(baseName))
			}
			err := ioutil.WriteFile(outputName, v, 0o644)
			if err != nil {
				log.Fatalf("writing output: %s", err)
			}
		}
	}
	return g
}

func Do(opts ...Option) {
	NewGenerator(opts...).ParsePackage().Generate().Format().Output()
}
