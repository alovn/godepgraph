package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ShowImports(root, showPkgName string, showStdLib, showThirdLib, isReverse bool) error {
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

	if isReverse { //reverse depencency
		//search pkg
		if showPkgName == "" {
			return errors.New("pkg name required")
		}
		fmt.Println(showPkgName)
		var selectDepPkgs []string
		for pkgName, depPkgs := range pkgMap {
			for depPkg := range depPkgs {
				if showPkgName == depPkg {
					// fmt.Println("  ", pkgName)
					selectDepPkgs = append(selectDepPkgs, pkgName)
				}
			}
		}
		for i, pkgName := range selectDepPkgs {
			flag := "├──"
			if i == len(selectDepPkgs)-1 {
				flag = "└──"
			}
			fmt.Println(flag, pkgName)
		}
		return nil
	}

	for pkgName, depPkgs := range pkgMap {
		if showPkgName != "" && pkgName != showPkgName {
			continue
		}
		fmt.Println(pkgName)
		var selectDepPkgs []string
		for depPkg, depPkgType := range depPkgs {
			if !showStdLib && depPkgType.PkgType == PkgTypeStandard {
				continue
			}
			if !showThirdLib && depPkgType.PkgType == PkgTypeThirdModule {
				continue
			}
			selectDepPkgs = append(selectDepPkgs, depPkg)
			// fmt.Println("  ", depPkg)
		}
		for i, depPkg := range selectDepPkgs {
			flag := "├──"
			if i == len(selectDepPkgs)-1 {
				flag = "└──"
			}
			fmt.Println(flag, depPkg)
		}
	}
	return nil
}

func ShowImportsWithGraphviz(root, showPkgName string, showStdLib, showThirdLib, reverse bool, output string) error {
	if output == "" {
		return errors.New("error: output")
	}
	format := strings.ToLower(strings.TrimPrefix(filepath.Ext(output), "."))
	supportFormat := map[string]bool{
		"png": true,
		"jpg": true,
		"gif": true,
		"svg": true,
		"dot": true,
	}
	if _, ok := supportFormat[format]; !ok {
		return errors.New("error: output format not support!")
	}
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
	if err := OutputGraphFormat(&builder, root, showPkgName, showStdLib, showThirdLib, reverse); err != nil {
		return err
	}
	if format == "dot" {
		file, err := os.Create(output)
		if err != nil {
			return errors.New("error create temp file")
		}
		defer file.Close()
		_, err = file.WriteString(builder.String())
		if err != nil {
			return fmt.Errorf("error write temp file: %v", err)
		}
		return nil
	} else {
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
		execCmd := exec.Command("dot", fmt.Sprintf("-T%s", format), tmpFilePath, fmt.Sprintf("-o%s", output))
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		// fmt.Println(execCmd.String())
		_ = execCmd.Run()
		return nil
	}
}
