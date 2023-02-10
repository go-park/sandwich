package gen

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-park/sandwich/pkg/aspect"
	"github.com/go-park/sandwich/pkg/astutils"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	options
	pkgList           map[string]*astutils.Package // Package we are scanning.
	aspectCache       map[string]aspect.Aspect
	aspectAlias       map[string]string
	aspectCustoms     map[astutils.Annotation]string
	proxyCache        map[*ast.Ident]aspect.Proxy
	delayAspectLoader map[astutils.Annotation][]func()
	componentCache    map[string]aspect.Component
}

func NewGenerator(opts ...Option) *Generator {
	ge := &Generator{
		options:           DefaultOptions(),
		pkgList:           map[string]*astutils.Package{},
		aspectCache:       map[string]aspect.Aspect{},
		aspectAlias:       map[string]string{},
		aspectCustoms:     map[astutils.Annotation]string{},
		proxyCache:        map[*ast.Ident]aspect.Proxy{},
		delayAspectLoader: map[astutils.Annotation][]func(){},
		componentCache:    map[string]aspect.Component{},
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
		item := &astutils.Package{
			Name:              pkg.Name,
			Path:              pkg.ID,
			Pwd:               getCurrentPkg(),
			AstPkg:            pkg,
			Defs:              map[*ast.Ident]types.Object{},
			Files:             make([]*astutils.File, len(pkg.Syntax)),
			OutputFiles:       map[string][]byte{},
			FileBuf:           map[string]bytes.Buffer{},
			AspectCache:       g.aspectCache,
			AspectAlias:       g.aspectAlias,
			AspectCustoms:     g.aspectCustoms,
			ProxyCache:        g.proxyCache,
			DelayAspectLoader: g.delayAspectLoader,
			ComponentCache:    g.componentCache,
		}
		for i, file := range pkg.Syntax {
			item.Files[i] = &astutils.File{
				File:    file,
				Pkg:     item,
				Imports: map[string]string{},
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
				ast.Inspect(file.File, file.InspectGenDecl)
			}
		}
	}
	for _, pkg := range g.pkgList {
		for _, file := range pkg.Files {
			if file.File != nil {
				ast.Inspect(file.File, file.InspectFuncDecl)
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
			log.Panic("unexported struct cannot be proxy")
		}
		abstract := proxy.Abstract()
		if len(abstract) == 0 {
			// parent = proxy.Name()
			abstract = "*" + proxy.Name() + proxy.Suffix()
		}
		pd := astutils.ProxyData{
			Package:         proxy.PkgName(),
			ProxyStructName: proxy.Name() + proxy.Suffix(),
			AbstractName:    abstract,
			ParentName:      proxy.Name(),
			Option:          proxy.Option(),
			Singleton:       proxy.IsSingleton(),
		}
		pd.Imports = append(pd.Imports, astutils.GetImports(proxy.Imports())...)
		for _, v := range proxy.Fields() {
			comp, ok := g.componentCache[v.Inject()]
			if !(ok || len(v.Assign()) > 0) {
				continue
			}
			var assign string
			if ok {
				facPkg, facPkgName, facName := comp.Factory()
				assign = facName + "()"
				if facPkg != proxy.PkgPath() {
					assign = facPkgName + "." + assign
				}
			}
			if len(v.Assign()) > 0 {
				assign = v.Assign()
			}
			pd.InjectFields = append(pd.InjectFields,
				&astutils.ProxyInjectField{
					Var: template.HTML(v.Name()),
					Val: template.HTML(assign),
				})
		}
		cuts := proxy.GetPointcuts()
		for _, method := range proxy.GetMethods() {
			cuts := append(cuts, method.GetPointcuts()...)
			paramNames, params := method.GetParams()
			resultNames, results := method.GetResults()
			m := &astutils.ProxyMethod{
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
				} else if anno, ok := g.aspectCustoms[astutils.Annotation(aspectName)]; ok {
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
		tpl, err := template.New("").Parse(astutils.GetProxyTpl())
		if err != nil {
			log.Panic(err.Error())
		}
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, pd); err != nil {
			log.Panic(err.Error())
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
				log.Println("output file:\n", v.String())
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
				targetPkg := getRelevantPkg(pkg.Pwd, pkg.Path)
				// not current project package
				if len(targetPkg) == 0 {
					continue
				}
				baseName := fmt.Sprintf("%s_proxy.gen.go", k)
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

var (
	buildTags string
	recursive bool
	deps      string
)

// Usage is a replacement usage function for the flags package.
func usage() {
	fmt.Fprintf(os.Stderr, "Usage of Aspect:\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "\thttps://github.com/go-park/sandwich\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func Do(opts ...Option) {
	log.SetFlags(0)
	log.SetPrefix("aspect: ")
	flag.StringVar(&buildTags, "tags", "", "comma-separated list of build tags to apply")
	flag.BoolVar(&recursive, "recursive", true, "true or false package load recursively, default true")
	flag.BoolVar(&recursive, "r", true, "true or false package load recursively, default true")
	flag.StringVar(&deps, "deps", "", "comma-separated list of dependencies need scan")

	flag.Usage = usage
	flag.Parse()

	opts = append([]Option{
		WithPatterns(flag.Args()...),
		WithRecursive(recursive),
		WithDeps(strings.Split(deps, ",")...),
		WithTags(strings.Split(buildTags, ",")...),
	}, opts...)

	NewGenerator(opts...).ParsePackage().Generate().Format().Output()
}
