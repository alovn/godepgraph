package app

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

func Serve(root string, addr string, showStdLib bool) error {
	mux := http.DefaultServeMux
	mux.HandleFunc("/graph", graphHandler(root, showStdLib))
	return http.ListenAndServe(addr, mux)
}

func graphHandler(root string, showStdLib bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if root == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return
			}
			root = cwd
		}
		module, err := ReadGoModule(root)
		if err != nil {
			return
		}
		pkgMap := make(map[string]map[string]PkgTypeInfo)
		if err := ReadDirImportPkgs(root, "", module, pkgMap); err != nil {
			return
		}
		var b bytes.Buffer
		fmt.Fprint(&b, `digraph godepgraph {
splines=curved
nodesep=0.5
ranksep=1.2
node [shape="box",style="rounded,filled"]
edge [arrowsize="0.5"]
`)

		for pkgName, depPkgs := range pkgMap {
			fmt.Fprintf(&b, "\"%s\" [label=\"%s\" color=\"green\" URL=\"https://pkg.go.dev/%s\" target=\"_blank\"];\n",
				pkgName,
				pkgName,
				module,
			)
			if len(depPkgs) > 0 {
				for depPkg, depPkgType := range depPkgs {
					if !showStdLib && depPkgType.PkgType == PkgTypeStandard {
						continue
					}
					fmt.Fprintf(&b, "\"%s\" -> %s;\n", pkgName, depPkg)
				}
			}
		}
		fmt.Fprintf(&b, "}")
		_, _ = w.Write(b.Bytes())
	}
}

func dotSplines(splines string) string {
	switch splines {
	case "line", "polyline", "curved", "ortho", "spline":
		return splines
	default:
		return "polyline"
	}
}

func dotNodeColor(pkgType PkgType) string {
	switch pkgType {
	case PkgTypeCurrentModule:
		return "green"
	case PkgTypeThirdModule:
		return "blue"
	case PkgTypeStandard:
		return "gray"
	default:
		return "white"
	}
}
