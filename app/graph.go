package app

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

var (
	cache PkgMap
)

func OutputGraphFormat(w io.Writer, root, showPkgName string, showStdLib bool) error {
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root = cwd
	}
	module, err := ReadGoModule(root)
	if err != nil {
		return err
	}

	//read cache first
	var pkgMap PkgMap
	if cache == nil {
		pkgMap = make(PkgMap)
		if err := ReadDirImportPkgs(root, "", module, pkgMap); err != nil {
			return err
		}
		cache = pkgMap
	} else {
		pkgMap = cache
	}

	var b bytes.Buffer
	fmt.Fprint(&b, `digraph godepgraph {
splines=curved
nodesep=0.8
ranksep=5
node [shape="box",style="rounded,filled"]
edge [arrowsize="0.8"]
`)
	labelsMap := make(map[string]string)
	for pkgName, depPkgs := range pkgMap {
		if showPkgName != "" && pkgName != showPkgName {
			continue
		}
		if _, ok := labelsMap[pkgName]; !ok {
			fmt.Fprintf(&b, "\"%s\" [label=\"%s\" fillcolor=\"white\" color=\"#0065FE\" fontcolor=\"#0065FE\" class=\"node_module\"];\n",
				pkgName,
				pkgName,
			)
			labelsMap[pkgName] = ""
		}

		if len(depPkgs) > 0 {
			for depPkg, depPkgType := range depPkgs {
				if !showStdLib && depPkgType.PkgType == PkgTypeStandard {
					continue
				}
				if depPkgType.PkgType == PkgTypeCurrentModule {
					for pkgName2 := range pkgMap { //depend packages label
						if pkgName2 != depPkg {
							continue
						}
						if _, ok := labelsMap[pkgName2]; ok {
							break
						}
						//label
						fmt.Fprintf(&b, "\"%s\" [label=\"%s\" fillcolor=\"white\" color=\"#0065FE\" fontcolor=\"#0065FE\" class=\"node_module\"];\n",
							pkgName2,
							pkgName2,
						)
					}
				}
				fmt.Fprintf(&b, "\"%s\" -> \"%s\";\n", pkgName, depPkg)
			}
		}
	}
	fmt.Fprintf(&b, "}")
	_, _ = w.Write(b.Bytes())
	return nil
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
