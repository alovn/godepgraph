package app

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type PkgMap map[string]map[string]PkgTypeInfo
type PkgTypeInfo struct {
	IsRoot  bool
	PkgType PkgType
}
type PkgType int

const (
	PkgTypeStandard PkgType = iota
	PkgTypeCurrentModule
	PkgTypeThirdModule
)

func ReadDirImportPkgs(rootPath, parentDirPath, module string, pkgMap PkgMap) error {
	dirPath := filepath.Join(rootPath, parentDirPath)
	dirs, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	var subDirs []string
	var goFiles []string
	for _, dir := range dirs {
		if strings.HasPrefix(dir.Name(), ".") {
			continue
		}
		if dir.IsDir() {
			subDirs = append(subDirs, dir.Name())
		} else {
			if filepath.Ext(dir.Name()) != ".go" {
				continue
			}
			if strings.HasSuffix(dir.Name(), "_test.go") {
				continue
			}
			goFiles = append(goFiles, filepath.Join(dirPath, dir.Name()))
		}
	}
	if len(goFiles) > 0 {
		if err := ReadGoFilesImportPkgs(rootPath, parentDirPath, module, goFiles, pkgMap); err != nil {
			return err
		}
	}
	for _, subDir := range subDirs {
		subDirPath := filepath.Join(parentDirPath, subDir)
		if err := ReadDirImportPkgs(rootPath, subDirPath, module, pkgMap); err != nil {
			return err
		}
	}
	return nil
}

func ReadGoFilesImportPkgs(rootPath, parentDirPath, module string, goFiles []string, pkgMap PkgMap) error {
	if len(goFiles) == 0 {
		return nil
	}
	for _, file := range goFiles {
		bs, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if len(bs) == 0 {
			continue
		}

		content := string(bs)
		pkgName, imports, err := parseImports(content)
		if err != nil {
			return err
		}

		var fullPkgName string
		if parentDirPath == "" {
			fullPkgName = "/"
		} else {
			if !strings.HasSuffix(parentDirPath, pkgName) { //pkg name not equals directory name
				fullPkgName = "/" + parentDirPath + "/" + pkgName
			} else {
				fullPkgName = "/" + parentDirPath
			}
		}

		for _, imp := range imports {
			var pkgType PkgType
			isRoot := imp == fmt.Sprintf("\"%s\"", filepath.Base(module)) || imp == fmt.Sprintf("\"%s\"", module)
			isCurrentModule := isRoot || strings.HasPrefix(imp, "\""+module)
			if isCurrentModule {
				if isRoot {
					imp = `"/"`
				} else {
					imp = strings.Replace(imp, module, "", 1)
				}
				pkgType = PkgTypeCurrentModule

			} else {
				isThirdModule := strings.Contains(imp, ".")
				if isThirdModule {
					pkgType = PkgTypeThirdModule
				} else {
					pkgType = PkgTypeStandard
				}
			}
			if pkgImports, ok := pkgMap[fullPkgName]; ok {
				pkgImports[imp] = PkgTypeInfo{
					IsRoot:  isRoot,
					PkgType: pkgType,
				}
			} else {
				pkgImports := make(map[string]PkgTypeInfo)
				pkgImports[imp] = PkgTypeInfo{
					IsRoot:  isRoot,
					PkgType: pkgType,
				}
				pkgMap[fullPkgName] = pkgImports
			}
		}
	}
	return nil
}

func parseImports(content string) (pkgName string, imports []string, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return
	}
	for _, imp := range f.Imports {
		imports = append(imports, imp.Path.Value)
	}
	return f.Name.String(), imports, nil
}

func ReadGoModule(root string) (string, error) {
	goModPath := filepath.Join(root, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}

		if strings.HasPrefix(line, "module") {
			// module := strings.TrimLeft(line, "module")
			module := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			return module, nil
		}
	}
	return "", nil
}
