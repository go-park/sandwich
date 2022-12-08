package astutils

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-park/sandwich/pkg/aspectlib"
)

var (
	regexParamTo    = regexp.MustCompile(`\.ParamTo\(([1-9][0-9]*)\)\.\((.*?)\)`)
	regexResultTo   = regexp.MustCompile(`\.ResultTo\(([1-9][0-9]*)\)\.\((.*?)\)`)
	regexAnnotation = regexp.MustCompile(`(@[A-Z][a-zA-Z]*)\(?.*\)?$`)
)

func trimQuotes(s string) string {
	return strings.TrimFunc(s, func(c rune) bool {
		return c == '"'
	})
}

func trimBrackets(s string) string {
	return strings.TrimFunc(s, func(c rune) bool {
		return c == '(' || c == ')'
	})
}

func getCommentParam(c *ast.CommentGroup, a aspectlib.Annotation) (ret map[aspectlib.AnnotationKey]string) {
	if c == nil {
		return
	}
	ret = make(map[aspectlib.AnnotationKey]string)
	for _, v := range strings.Split(c.Text(), "\n") {
		if strings.HasPrefix(v, a.String()) {
			str := strings.TrimPrefix(v, a.String())
			str = strings.TrimSpace(str)
			str = trimBrackets(str)
			str = strings.TrimSpace(str)
			for _, v := range strings.Split(str, ",") {
				v = strings.TrimSpace(v)
				if kv := strings.Split(v, `=`); len(kv) == 2 {
					key := aspectlib.AnnotationKey(kv[0])
					if aspectlib.IsSystemAnnotationKey(key) {
						ret[key] = trimQuotes(kv[1])
					}
					continue
				}
				v = trimQuotes(v)
				ret[aspectlib.CommentKeyDefault] = v
			}
		}
	}
	return ret
}

func GetImports(specs []*ast.ImportSpec) []*aspectlib.ProxyImport {
	var imports []*aspectlib.ProxyImport
	for _, v := range specs {
		imp := &aspectlib.ProxyImport{
			Alias: template.HTML(v.Name.String()),
			Path:  template.HTML(v.Path.Value),
		}
		if v.Name == nil {
			imp.Alias = ""
		}
		imports = append(imports, imp)
	}
	return imports
}

func replaceParamPlaceholder(advice aspectlib.Advice, method aspectlib.Method, stmt string) string {
	var jpName string
	var resultName string
	if advice.Func().Type.Params != nil && len(advice.Func().Type.Params.List) > 0 {
		param := advice.Func().Type.Params.List[0]
		if len(param.Names) > 0 {
			jpName = param.Names[0].Name
		}
	}
	if advice.Func().Type.Results != nil && len(advice.Func().Type.Results.List) > 0 {
		result := advice.Func().Type.Results.List[0]
		if len(result.Names) > 0 {
			resultName = result.Names[0].Name
		}
	}
	// replace function name placeholder
	funcStmt := jpName + ".FuncName()"
	stmt = strings.ReplaceAll(stmt, funcStmt, fmt.Sprintf(`"%s"`, method.Name()))

	// replace function args placeholder
	paramNames, params := method.GetParams()
	argsStmt := jpName + ".Params()"
	args := fmt.Sprintf("[]interface{}{%s}", strings.Join(paramNames, ", "))
	stmt = strings.ReplaceAll(stmt, argsStmt, args)

	// replace function result placeholder
	resultNames, results := method.GetResults()
	argsStmt = jpName + ".Results()"
	args = fmt.Sprintf("[]interface{}{%s}", strings.Join(resultNames, ", "))
	stmt = strings.ReplaceAll(stmt, argsStmt, args)

	// replace result assign placeholder
	for i, v := range resultNames {
		resultAssignStmt := fmt.Sprintf(resultName+"[%d]", i+1)
		stmt = strings.ReplaceAll(stmt, resultAssignStmt, v)
	}

	// replace param assert placeholder
	paramToStmt := jpName + ".ParamTo(%s).(%s)"
	for _, sub := range regexParamTo.FindAllStringSubmatchIndex(stmt, -1) {
		if len(sub) > 0 && strings.HasSuffix(stmt[0:sub[0]], jpName) {
			index := stmt[sub[2]:sub[3]]
			i, err := strconv.Atoi(index)
			if err != nil {
				panic(err)
			}
			typ := stmt[sub[4]:sub[5]]
			param := params[i-1]
			if strings.HasSuffix(param, typ) {
				paramName := strings.Split(strings.Split(param, " ")[0], ",")[0]
				raw := fmt.Sprintf(paramToStmt, index, typ)
				stmt = strings.Replace(stmt, raw, paramName, 1)
			}
		}
	}

	// replace result assert placeholder
	resultToStmt := jpName + ".ResultTo(%s).(%s)"
	for _, sub := range regexResultTo.FindAllStringSubmatchIndex(stmt, -1) {
		if len(sub) > 0 && strings.HasSuffix(stmt[0:sub[0]], jpName) {
			index := stmt[sub[2]:sub[3]]
			i, err := strconv.Atoi(index)
			if err != nil {
				panic(err)
			}
			typ := stmt[sub[4]:sub[5]]
			result := results[i-1]
			if strings.HasSuffix(result, typ) {
				resultName := strings.Split(strings.Split(result, " ")[0], ",")[0]
				raw := fmt.Sprintf(resultToStmt, index, typ)
				stmt = strings.Replace(stmt, raw, resultName, 1)
			}
		}
	}

	// replace invalid assignment
	if l := strings.Split(stmt, ":="); len(l) == 2 {
		if strings.TrimSpace(l[0]) == strings.TrimSpace(l[1]) {
			stmt = "-"
		}
	}

	// replace proceed placeholder
	proceedStmt := jpName + ".Proceed"
	if strings.Contains(stmt, proceedStmt) {
		stmt = "-proceed"
	}

	// return statement add prefix
	if stmt == "return" || strings.HasPrefix(stmt, "return ") {
		stmt = "-" + stmt
	}
	return stmt
}

func ParseAdviceStmt(advice aspectlib.Advice, method aspectlib.Method) []string {
	var list []string
	if advice == nil || advice.Func() == nil {
		return list
	}
	for _, stmt := range advice.Func().Body.List {
		var buf bytes.Buffer
		_ = printer.Fprint(&buf, token.NewFileSet(), stmt)
		s := strings.TrimSpace(buf.String())
		for _, v := range strings.Split(s, "\n\t") {
			for _, v := range strings.Split(v, "\n") {
				s = replaceParamPlaceholder(advice, method, v)
				list = append(list, s)
			}
		}
	}
	return list
}

func ParseAroundAdvice(advice aspectlib.Advice, method aspectlib.Method) ([]string, []string) {
	var before, after []string
	stmt := ParseAdviceStmt(advice, method)
	stmtLen := len(stmt)
	proceedIndex := len(stmt) + 1
	for i, v := range stmt {
		if v == "-proceed" {
			proceedIndex = i + 1
		}
		if strings.HasPrefix(v, "-return") {
			stmt[i] = strings.TrimPrefix(v, "-")
		}
	}
	lastStmt := stmt[stmtLen-1]
	if lastStmt == "return" || strings.HasPrefix(lastStmt, "return ") {
		stmt[stmtLen-1] = "-" + lastStmt
	}
	before = stmt[:proceedIndex:proceedIndex]
	if proceedIndex < len(stmt)+1 {
		after = stmt[proceedIndex:]
	}
	return before, after
}

func parseAnnotation(c *ast.CommentGroup) []aspectlib.Annotation {
	if c == nil {
		return nil
	}
	var result []aspectlib.Annotation
	for _, v := range strings.Split(c.Text(), "\n") {
		if ss := regexAnnotation.FindStringSubmatch(v); len(ss) > 1 {
			anno := aspectlib.Annotation(ss[1])
			result = append(result, anno)
		}
	}
	return result
}

func validCustomAnnotation(name string) (aspectlib.Annotation, bool) {
	full := "@" + name
	if regexAnnotation.MatchString(full) {
		anno := aspectlib.Annotation(full)
		if !aspectlib.IsSystemAnnotation(anno) && len(full) > 1 {
			return anno, true
		}
	}
	return "", false
}
