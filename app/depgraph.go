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

var (
	pkgMap = make(map[string][]string)
)

type PkgMap map[string]map[string]string

func Imports(root string) error {
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root = cwd
	}
	// root = "/Users/alovn/workspace/github/gostack-labs/bytego"
	// module, err := ReadGoModule(root)
	// if err != nil {
	// 	return err
	// }
	// log.Println(module)
	pkgMap := make(map[string]map[string]string)
	if err := ReadDirImportPkgs(root, "", pkgMap); err != nil {
		return err
	}

	for k, v := range pkgMap {
		fmt.Println(k)
		if len(v) > 0 {
			for k2 := range v {
				fmt.Println("  ", k2)
			}
		}
	}
	return nil
}

func ReadDirImportPkgs(rootPath string, parentDirPath string, pkgMap PkgMap) error {
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
		if err := ReadGoFilesImportPkgs(rootPath, parentDirPath, goFiles, pkgMap); err != nil {
			return err
		}
	}
	for _, subDir := range subDirs {
		subDirPath := filepath.Join(parentDirPath, subDir)
		if err := ReadDirImportPkgs(rootPath, subDirPath, pkgMap); err != nil {
			return err
		}
	}
	return nil
}

func ReadGoFilesImportPkgs(rootPath string, parentDirPath string, goFiles []string, pkgMap PkgMap) error {
	if len(goFiles) == 0 {
		return nil
	}
	// var pkgName string
	for _, file := range goFiles {
		// pkgName = file
		bs, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if len(bs) == 0 {
			continue
		}

		content := string(bs)
		pkgName, imports, err := parseImports(content)
		fullPkgName := parentDirPath
		if !strings.HasSuffix(parentDirPath, pkgName) {
			fullPkgName = filepath.Join(parentDirPath, pkgName)
		}
		if err != nil {
			return err
		}
		if pkgImports, ok := pkgMap[fullPkgName]; ok {
			for _, imp := range imports {
				pkgImports[imp] = ""
			}
		} else {
			pkgImports := make(map[string]string)
			for _, imp := range imports {
				pkgImports[imp] = ""
			}
			pkgMap[fullPkgName] = pkgImports
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
			module := strings.Trim(strings.TrimPrefix(line, "module"), " ")
			return module, nil
		}
	}
	return "", nil
}
