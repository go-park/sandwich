package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-park/sandwich/pkg/gen"
)

var (
	buildTags string
	recursive bool
	deps      string
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of Aspect:\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "\thttps://github.com/go-park/sandwich\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("aspect: ")
	flag.StringVar(&buildTags, "tags", "", "comma-separated list of build tags to apply")
	flag.BoolVar(&recursive, "recursive", true, "true or false package load recursively, default true")
	flag.BoolVar(&recursive, "r", true, "true or false package load recursively, default true")
	flag.StringVar(&deps, "deps", "", "comma-separated list of dependencies need scan")

	flag.Usage = Usage
	flag.Parse()

	gen.Do(
		gen.WithPatterns(flag.Args()...),
		gen.WithRecursive(recursive),
		gen.WithDeps(strings.Split(deps, ",")...),
		gen.WithTags(strings.Split(buildTags, ",")...),
	)
}
