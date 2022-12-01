package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func getAllPathPatterns(patterns []string) []string {
	var list []string
	for _, v := range patterns {
		_ = filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return nil
			}
			if info.IsDir() {
				list = append(list, strings.Join([]string{".", path}, string(filepath.Separator)))
			}
			return nil
		})
	}
	return list
}

func getCurrentPkg() string {
	pkgs, _ := packages.Load(&packages.Config{}, ".")
	return pkgs[0].String()
}

func getRelevantPkg(current, target string) (s string) {
	// target deep than current, trim current
	if strings.HasPrefix(target, current) {
		s = strings.TrimPrefix(target, current)
	} else {
		var prefix string
		list := strings.Split(current, string(filepath.Separator))
		// current deep than targe, get layer
		for _, p := range strings.Split(target, string(filepath.Separator)) {
			prepre := prefix

			prefix = filepath.Join(prefix, p)
			if !strings.HasPrefix(current, prefix) {
				prefix = prepre
				break
			}
			list = list[1:]
		}
		s = strings.TrimPrefix(target, prefix)
		for range list {
			s = filepath.Join("..", s)
		}
	}
	s = filepath.Join(".", s)
	return
}

func filterEmptyStr(ss ...string) []string {
	arr := make([]string, 0, len(ss))
	for _, s := range ss {
		if len(s) > 0 {
			arr = append(arr, s)
		}
	}
	return arr
}
