package astutils

import (
	"go/ast"

	"github.com/go-park/sandwich/pkg/aspect"
	"github.com/go-park/sandwich/pkg/tools/collections"
)

type (
	ProxyInterceptor func([]Annotation, aspect.Proxy, *ast.StructType) []aspect.ProxyOption
	FieldInterceptor func([]Annotation, aspect.Field, *ast.Field) []aspect.FieldOption
)

func init() {
	proxyInterceptors = append(proxyInterceptors, proxyParamProcess)
}

var (
	proxyInterceptors []ProxyInterceptor
	fieldInterceptors []FieldInterceptor
)

func RegisterProxyInterceptors(opts ...ProxyInterceptor) {
	proxyInterceptors = append(proxyInterceptors, opts...)
}

func RegisterFieldInterceptors(opts ...FieldInterceptor) {
	fieldInterceptors = append(fieldInterceptors, opts...)
}

func proxyParamProcess(ann []Annotation, pro aspect.Proxy, t *ast.StructType) (result []aspect.ProxyOption) {
	params := GetCommentParam(pro.Docs(), CommentProxy)
	// proxy object
	abstract := params[CommentKeyAbstract]
	if v, ok := params[CommentKeyDefault]; ok {
		abstract = v
	}
	// factory method suffix
	suffix := DefaultProxySuffix
	if v, ok := params[CommentKeySuffix]; ok {
		suffix = v
	}
	// factory method option
	option := params[CommentKeyOption]
	// is singleton
	var singleton bool
	s := params[CommentKeySingleton]
	if s == "true" {
		singleton = true
	}

	result = append(result,
		aspect.WithProxyAbstract(abstract),
		aspect.WithProxySuffix(suffix),
		aspect.WithProxyOption(option),
		aspect.WithProxyMode(singleton),
	)
	var pos []aspect.Pointcut
	if collections.Contains(ann, CommentPointcut) {
		params := GetCommentParam(pro.Docs(), CommentPointcut)
		for _, v := range params {
			pos = append(pos, aspect.NewPointcut(aspect.WithPointcutName(v)))
		}
	}
	result = append(result, aspect.WithProxyPointcuts(pos...))
	return
}
