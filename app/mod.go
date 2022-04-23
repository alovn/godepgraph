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

type ModuleGraphNode struct {
	Name    string
	PkgType PkgType
	Deps    []*ModuleGraphNode
	Parent  *ModuleGraphNode `json:"-"`
	Index   int
	Depth   int
}

type ModGraphLink struct {
	Source string
	Target string
}

func buildModGraphMap(path string) (pkgMap PkgMap, links []ModGraphLink, err error) {
	execCmd := exec.Command("go", "mod", "graph")
	execCmd.Dir = path
	execCmd.Stderr = os.Stderr
	// execCmd.Stdout = os.Stdout
	var out bytes.Buffer
	execCmd.Stdout = &out
	if err = execCmd.Run(); err != nil {
		return
	}
	pkgMap = make(PkgMap)
	reader := bufio.NewReader(&out)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
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
		links = append(links, ModGraphLink{
			Source: source,
			Target: target,
		})

	}
	return
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
		_, _ = w.Write([]byte(fmt.Sprintf("%s%s\n", prefixStr+flag, dep.Name)))
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

func outputModGraph(w io.Writer, path, rootModule, findModule string, isReverse, isTreeStyle bool) error {
	pkgMap, links, err := buildModGraphMap(path)
	if err != nil {
		return err
	}
	if findModule == "" {
		findModule = rootModule
	}
	if isReverse {
		return OutputModGraphReverse(w, links, findModule, isTreeStyle)
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

	if isTreeStyle {
		fmt.Println(findModule)
		outputModGraphTree(w, tree, 0)
	} else {

		outputModGraphViz(w, tree)
	}
	return nil
}

func OutputModGraphReverse(w io.Writer, links []ModGraphLink, findModule string, isTree bool) error {
	var buf bytes.Buffer
	if isTree {
		fmt.Fprintln(&buf, findModule)
		for _, link := range links {
			if link.Target == findModule {
				fmt.Fprintln(&buf, link.Source)
			}
		}
	} else {
		fmt.Fprint(&buf, `digraph godepgraph {
			splines=curved
			nodesep=0.8
			ranksep=5
			node [shape="box",style="rounded,filled"]
			edge [arrowsize="0.8"]
			`)
		for _, link := range links {
			if link.Target == findModule {
				fmt.Fprintf(&buf, "\"%s\" -> \"%s\";\n", link.Source, link.Target)
			}
		}
		fmt.Fprintf(&buf, "}")
	}
	_, _ = w.Write(buf.Bytes())
	return nil
}

func outputModGraphViz(w io.Writer, tree *ModuleGraphNode) {
	var buf bytes.Buffer
	hashMap := make(map[string]bool)
	fmt.Fprint(&buf, `digraph godepgraph {
splines=curved
nodesep=0.8
ranksep=5
node [shape="box",style="rounded,filled"]
edge [arrowsize="0.8"]
`)
	outputModGraphNode(&buf, tree, 0, hashMap)
	fmt.Fprintf(&buf, "}")
	_, _ = w.Write(buf.Bytes())
}

func outputModGraphNode(b *bytes.Buffer, tree *ModuleGraphNode, depth int, hashMap map[string]bool) {
	if tree == nil {
		return
	}
	depth++
	for _, dep := range tree.Deps {
		var key string
		if n := strings.Compare(tree.Name, dep.Name); n > -1 {
			key = tree.Name + dep.Name
		} else {
			key = dep.Name + tree.Name
		}
		if _, ok := hashMap[key]; ok {
			continue
		}
		hashMap[key] = true
		fmt.Fprintf(b, "\"%s\" -> \"%s\";\n", tree.Name, dep.Name)
		outputModGraphNode(b, dep, depth, hashMap)
	}
}
