package astutils

import (
	"go/ast"

	"github.com/go-park/sandwich/pkg/aspect"
)

type (
	ProxyInterceptor func([]Annotation, aspect.Proxy, *ast.StructType) []aspect.ProxyOption
	FieldInterceptor func([]Annotation, aspect.Field, *ast.Field) []aspect.FieldOption
)

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
