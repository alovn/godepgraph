package app

import (
	"fmt"
	"os"
)

func ShowImports(root string, showStdLib bool) error {
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root = cwd
	}
	// root = "/Users/alovn/workspace/github/gostack-labs/bytego"
	module, err := ReadGoModule(root)
	if err != nil {
		return err
	}
	pkgMap := make(map[string]map[string]PkgTypeInfo)
	if err := ReadDirImportPkgs(root, "", module, pkgMap); err != nil {
		return err
	}

	for pkgName, depPkgs := range pkgMap {
		fmt.Println(pkgName)
		if len(depPkgs) > 0 {
			for depPkg, depPkgType := range depPkgs {
				if !showStdLib && depPkgType.PkgType == PkgTypeStandard {
					continue
				}
				fmt.Println("  ", depPkg)
			}
		}
	}
	return nil
}
