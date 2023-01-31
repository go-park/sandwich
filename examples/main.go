package main

import (
	"go/ast"
	"strings"

	"github.com/go-park/sandwich/pkg/aspect"
	"github.com/go-park/sandwich/pkg/astutils"
	"github.com/go-park/sandwich/pkg/gen"
	"github.com/go-park/sandwich/pkg/tools/collections"
)

//go:generate go run ./... .
func main() {
	astutils.RegisterFieldInterceptors(ValueInterceotor)
	gen.Do()
}

func ValueInterceotor(list []astutils.Annotation, fi aspect.Field, f *ast.Field) (results []aspect.FieldOption) {
	valueAnno := astutils.Annotation("@Value")
	if collections.Contains(list, valueAnno) {
		params := astutils.GetCommentParam(f.Doc, valueAnno)
		if v, ok := params[astutils.CommentKeyDefault]; ok {
			if strings.Split(fi.Define(), " ")[1] == "string" {
				v = `"` + v + `"`
			}
			results = append(results, aspect.WithFieldAssign(v))
		}
	}
	return
}
