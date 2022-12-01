package gen

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getAllPathPatterns(t *testing.T) {
	type args struct {
		patterns []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAllPathPatterns(tt.args.patterns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllPathPatterns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCurrentPkg(t *testing.T) {
	assert.Equal(t, "github.com/go-park/sandwich/pkg/gen", getCurrentPkg())
}
