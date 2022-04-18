package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ShowImports(root string, showStdLib bool) error {
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
	if module == "" {
		return errors.New("error: the path must be a go module directory.")
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

func ShowImportsWithGraphviz(root string, showStdLib bool) error {
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
	if module == "" {
		return errors.New("error: the path must be a go module directory.")
	}
	pkgMap := make(map[string]map[string]PkgTypeInfo)
	if err := ReadDirImportPkgs(root, "", module, pkgMap); err != nil {
		return err
	}
	var builder strings.Builder
	if err := OutputGraphFormat(&builder, root, showStdLib); err != nil {
		return err
	}
	file, err := os.CreateTemp("", "godepgraph-*.dot")
	if err != nil {
		return errors.New("error create temp file")
	}
	tmpFilePath := file.Name()
	defer os.Remove(tmpFilePath)
	_, err = file.WriteString(builder.String())
	if err != nil {
		return fmt.Errorf("error write temp file: %v", err)
	}
	_ = file.Close()
	execCmd := exec.Command("dot", "-Tpng", tmpFilePath, "-o godepgraph.png")
	return execCmd.Run()
}
