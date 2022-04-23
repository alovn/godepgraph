package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func OutputDepGraph(path, findPkgName string, isShowStdLib, isShowThirdLib, isReverse, isModGraph bool) error {
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
	if module == "" {
		return errors.New("error: the path must be a go module directory.")
	}
	if isModGraph { //go mod graph tree
		return outputModGraph(os.Stdout, path, module, findPkgName, isReverse, true)
	} else {
		return outputPkgImports(path, module, findPkgName, isShowStdLib, isShowThirdLib, isReverse)
	}
}

func outputPkgImports(path, module, findPkgName string, isShowStdLib, isShowThirdLib, isReverse bool) error {
	pkgMap := make(PkgMap)
	if err := ReadDirImportPkgs(path, "", module, pkgMap); err != nil {
		return err
	}
	if isReverse { //reverse depencency
		//search pkg
		if findPkgName == "" {
			return errors.New("pkg name required")
		}
		fmt.Println(findPkgName)
		var selectDepPkgs []string
		for pkgName, depPkgs := range pkgMap {
			for depPkg := range depPkgs {
				if findPkgName == depPkg {
					selectDepPkgs = append(selectDepPkgs, pkgName)
				}
			}
		}
		for i, pkgName := range selectDepPkgs {
			flag := "├──"
			if i == len(selectDepPkgs)-1 {
				flag = "└──"
			}
			fmt.Printf("%s%s\n", flag, pkgName)
		}
		return nil
	}

	for pkgName, depPkgs := range pkgMap {
		if findPkgName != "" && pkgName != findPkgName {
			continue
		}
		fmt.Println(pkgName)
		var selectDepPkgs []string
		for depPkg, depPkgType := range depPkgs {
			if !isShowStdLib && depPkgType.PkgType == PkgTypeStandard {
				continue
			}
			if !isShowThirdLib && depPkgType.PkgType == PkgTypeThirdModule {
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
			fmt.Printf("%s%s\n", flag, depPkg)
		}
	}
	return nil
}

func OutputDepGraphviz(path, findPkgName string, isShowStdLib, isShowThirdLib, isReverse, isModGraph bool, output string) error {
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
	if module == "" {
		return errors.New("error: the path must be a go module directory.")
	}

	var builder strings.Builder
	if err := OutputGraphFormat(&builder, path, findPkgName, isShowStdLib, isShowThirdLib, isReverse, isModGraph); err != nil {
		return err
	}

	if format == "dot" { //output dot file
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
	} else { //output a picture
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
