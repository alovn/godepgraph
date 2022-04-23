/*
Copyright © 2022 alovn <alovn@live.com>
*/
package cmd

import (
	"os"

	"github.com/alovn/godepgraph/app"
	"github.com/spf13/cobra"
)

var (
	path        string
	pkg         string
	web         bool
	listen      string
	isShowStd   bool
	isShowThird bool
	isReverse   bool
	isModGraph  bool
	dot         bool
	output      string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "godepgraph",
	Short: "Go packages dependence packages graph",
	Long: `godepgraph is a tools for Go that dependence packages show with graph:

godepgraph
godepgraph --pkg=bytego
godepgraph --pkg=bytego --reverse
godepgraph --mod
godepgraph --web
godepgraph --std --third
godepgraph --dot
godepgraph --dot --output=/path/godepgraph.png
godepgraph --web --pkg=bytego
godepgraph --path=./myapp/ --pkg=bytego --web --listen=:7788`,
	Run: func(cmd *cobra.Command, args []string) {
		write := func(s string) {
			_, _ = os.Stderr.WriteString(s)
		}
		if isReverse && pkg == "" {
			write("the reverse need the parameter --pkg")
			return
		}
		if web {
			if err := app.Serve(path, listen, pkg, isShowStd, isShowThird, isReverse); err != nil {
				write(err.Error())
				return
			}
		} else {
			if dot {
				if err := app.OutputDepGraphviz(path, pkg, isShowStd, isShowThird, isReverse, isModGraph, output); err != nil {
					write(err.Error())
					return
				}
			} else if err := app.OutputDepGraph(path, pkg, isShowStd, isShowThird, isReverse, isModGraph); err != nil {
				write(err.Error())
				return
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&path, "path", ".", "the local path of packages")
	rootCmd.Flags().StringVar(&pkg, "pkg", "", "the go package namge")
	rootCmd.Flags().BoolVar(&web, "web", false, "serve a local web server and show the depgraph in the webpage")
	rootCmd.Flags().StringVar(&listen, "listen", "localhost:7788", "listen address of web server")
	rootCmd.Flags().BoolVar(&isShowStd, "std", false, "show std lib")
	rootCmd.Flags().BoolVar(&isShowThird, "third", false, "ishow third lib")
	rootCmd.Flags().BoolVar(&isReverse, "reverse", false, "reverse dependency")
	rootCmd.Flags().BoolVar(&isModGraph, "mod", false, "go mod graph")
	rootCmd.Flags().BoolVar(&dot, "dot", false, "generate a picture using graphviz")
	rootCmd.Flags().StringVar(&output, "output", "./godepgraph.png", "the output path of picture, supoort format:jpg,png,svg,gif,dot")
}
