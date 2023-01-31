package astutils

import "html/template"

type ProxyData struct {
	Package         string
	Imports         []*ProxyImport
	ProxyStructName string
	Methods         []*ProxyMethod
	AbstractName    string
	ParentName      string
	InjectFields    []*ProxyInjectField
}

type ProxyMethod struct {
	Name        string
	Params      string
	ParamNames  string
	Results     string
	ResultNames string
	Before      []any
	After       []any
}

type ProxyImport struct {
	Alias template.HTML
	Path  template.HTML
}

type ProxyInjectField struct {
	Var template.HTML
	Val template.HTML
}

const proxyTpl = `
// Code generated by sandwich. DO NOT EDIT.

package {{.Package}}

import (
	{{- range $i, $s := .Imports }}
	{{ $s.Alias}} {{ $s.Path}}
	{{- end}}
)

type {{ .ProxyStructName }} struct {
	parent {{ .ParentName }}
}

//@Component
func New{{ .ProxyStructName }}() {{ .AbstractName }} {
	return &{{ .ParentName }}{
	{{- range $i, $a := .InjectFields }}
	{{ $a.Var }}: {{ $a.Val }},
	{{- end }}
	}
}

{{ range .Methods }}
func (p *{{$.ProxyStructName}}) {{ .Name }} ({{ .Params }}) ({{ .Results }}) {
	{{- range $i, $s := .Before }}
	{{ $s }}
	{{- end }}
	{{- range $i, $s := .After }}
	{{ $s }}
	{{- end }}
	return {{ .ResultNames }}
}
{{ end }}
`

func GetProxyTpl() string {
	return proxyTpl
}

const (
	DefaultProxySuffix = "Proxy"
)
