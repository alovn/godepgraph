package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	cache PkgMap
)

func OutputGraphFormat(w io.Writer, path, findPkgName string, isShowStdLib, isShowThirdLib, isReverse, modGraph bool) error {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	}
	module, err := ReadGoModule(path)
	if err != nil {
		return err
	}
	if modGraph {
		return outputModGraph(w, path, module, findPkgName, isReverse, false)
	} else {
		return outputPkgGraph(w, path, module, findPkgName, isShowStdLib, isShowThirdLib, isReverse, modGraph)
	}
}

func outputPkgGraph(w io.Writer, path, module, findPkgName string, isShowStdLib, isShowThirdLib, isReverse, modGraph bool) error {
	//read cache first
	var pkgMap PkgMap
	if cache == nil {
		pkgMap = make(PkgMap)
		if err := ReadDirImportPkgs(path, "", module, pkgMap); err != nil {
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
	//reverse depencency
	if isReverse {
		//search pkg
		if findPkgName == "" {
			return errors.New("pkg name required")
		}
		fmt.Fprintf(&b, "\"%s\" [label=\"%s\" fillcolor=\"white\" color=\"#0065FE\" fontcolor=\"#0065FE\" class=\"node_module\"];\n",
			findPkgName,
			findPkgName,
		)
		for pkgName, depPkgs := range pkgMap {
			for depPkg := range depPkgs {
				if findPkgName == depPkg {
					fmt.Fprintf(&b, "\"%s\" [label=\"%s\" fillcolor=\"white\" color=\"#0065FE\" fontcolor=\"#0065FE\" class=\"node_module\"];\n",
						pkgName,
						pkgName,
					)
					fmt.Fprintf(&b, "\"%s\" -> \"%s\";\n", pkgName, depPkg)
				}
			}
		}
	} else {
		//normal
		labelsMap := make(map[string]string)
		for pkgName, depPkgs := range pkgMap {
			if findPkgName != "" && pkgName != findPkgName {
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
					if !isShowStdLib && depPkgType.PkgType == PkgTypeStandard {
						continue
					}
					if !isShowThirdLib && depPkgType.PkgType == PkgTypeThirdModule {
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
							labelsMap[pkgName2] = ""
						}
					}
					fmt.Fprintf(&b, "\"%s\" -> \"%s\";\n", pkgName, depPkg)
				}
			}
		}
	}
	fmt.Fprintf(&b, "}")
	_, _ = w.Write(b.Bytes())
	return nil
}
