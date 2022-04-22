package app

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func buildModGraphMap(path string) (PkgMap, error) {
	execCmd := exec.Command("go", "mod", "graph")
	execCmd.Dir = path
	execCmd.Stderr = os.Stderr
	// execCmd.Stdout = os.Stdout
	var out bytes.Buffer
	execCmd.Stdout = &out
	err := execCmd.Run()
	if err != nil {
		return nil, err
	}
	pkgMap := make(PkgMap)
	reader := bufio.NewReader(&out)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		arr := strings.Split(string(line), " ")
		if len(arr) != 2 {
			continue
		}
		source, target := arr[0], arr[1]
		deps, ok := pkgMap[source]
		if ok {
			deps[target] = PkgTypeInfo{
				IsRoot:  false,
				PkgType: PkgTypeThirdModule,
			}
		} else {
			pkgMap[source] = map[string]PkgTypeInfo{
				target: {
					IsRoot:  false,
					PkgType: PkgTypeThirdModule,
				},
			}
		}

	}
	return pkgMap, nil
}

type ModuleGraphNode struct {
	Name    string
	PkgType PkgType
	Deps    []*ModuleGraphNode
	Parent  *ModuleGraphNode `json:"-"`
	Index   int
	Depth   int
}

func buildModGraphTree(module string, allPkgMap PkgMap, parent *ModuleGraphNode, depth int) error {
	if deps, ok := allPkgMap[module]; ok {
		for depModule := range deps {
			// circle check
			if isCircle := nodeCircleCheck(parent, depModule); isCircle {
				continue
			}
			node := &ModuleGraphNode{
				Name:    depModule,
				PkgType: PkgTypeThirdModule,
				Parent:  parent,
				Index:   0,
				Depth:   parent.Depth + 1,
			}
			err := buildModGraphTree(depModule, allPkgMap, node, depth+1)
			if err != nil {
				return err
			}
			parent.Deps = append(parent.Deps, node)
		}
	}
	return nil
}

func outputModGraphTree(w io.Writer, tree *ModuleGraphNode, depth int) {
	if tree == nil {
		return
	}
	depth++
	if tree.Deps == nil {
		return
	}

	for i, dep := range tree.Deps {
		flag := "├──"
		isEnd := i == len(tree.Deps)-1
		if isEnd {
			flag = "└──"
		}
		var prefixStr string
		if dep.Depth > 1 {
			prefix := getPrintPrefix(tree.Parent)
			prefixStr = strings.Join(prefix, "")
		}
		_, _ = w.Write([]byte(fmt.Sprintf("%s %s\n", prefixStr+flag, dep.Name)))
		// fmt.Println(prefixStr+flag, dep.Name)
		tree.Index++
		outputModGraphTree(w, dep, depth)
	}
}

func nodeCircleCheck(node *ModuleGraphNode, name string) bool {
	if node == nil {
		return false
	}
	if node.Name == name {
		return true
	}

	if node.Parent != nil {
		return nodeCircleCheck(node.Parent, name)
	}
	return false
}

func getPrintPrefix(node *ModuleGraphNode) []string {
	var prefix []string
	if node == nil {
		return prefix
	}

	var tag string
	if node.Index < len(node.Deps) {
		tag = "│  "

	} else {
		tag = "   "
	}

	// tag = fmt.Sprintf("%d  ", num)
	// prefix = append(prefix, "")
	// copy(prefix[1:], prefix[0:])
	// prefix[0] = tag
	prefix = append([]string{tag}, prefix...)
	// prefix = append(prefix, tag)

	if node.Parent != nil {
		prefix2 := getPrintPrefix(node.Parent)
		prefix = append(prefix2, prefix...)
	}
	return prefix
}

func OutputModGraph(w io.Writer, path, rootModule, findModule string, isTree bool) error {
	pkgMap, err := buildModGraphMap(path)
	if err != nil {
		return err
	}
	if findModule == "" {
		findModule = rootModule
	}
	depth := 0
	tree := &ModuleGraphNode{
		Name:    findModule,
		PkgType: PkgTypeThirdModule,
		Parent:  nil,
		Index:   0,
		Depth:   0,
	}
	err = buildModGraphTree(findModule, pkgMap, tree, depth)
	if err != nil {
		return err
	}

	if isTree {
		fmt.Println(findModule)
		outputModGraphTree(w, tree, 0)
	} else {
		outputModGraphViz(w, tree)
	}
	return nil
}

func outputModGraphViz(w io.Writer, tree *ModuleGraphNode) {
	var b bytes.Buffer
	fmt.Fprint(&b, `digraph godepgraph {
splines=curved
nodesep=0.8
ranksep=5
node [shape="box",style="rounded,filled"]
edge [arrowsize="0.8"]
`)
	outputModGraphNode(&b, tree, 0)
	fmt.Fprintf(&b, "}")
	_, _ = w.Write(b.Bytes())
}

func outputModGraphNode(b *bytes.Buffer, tree *ModuleGraphNode, depth int) {
	if tree == nil {
		return
	}
	depth++
	for _, dep := range tree.Deps {
		fmt.Fprintf(b, "\"%s\" -> \"%s\";\n", tree.Name, dep.Name)
		fmt.Println()
		outputModGraphNode(b, dep, depth)
	}
}
