/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"os"

	"github.com/alovn/godepgraph/app"
	"github.com/spf13/cobra"
)

var (
	path  string = "."
	web   bool   = false
	port  int    = 7788
	depth int    = 5
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "godepgraph",
	Short: "Go packages dependence graph",
	Long: `godepgraph is a tools for Go that show dependence graph:

godepgraph
godepgraph --path=./myapp/
godepgraph --path=./myapp/ --web
godepgraph --path=./myapp/ --web --port=7788`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if web {

		} else {
			if err := app.Imports(path); err != nil {
				log.Fatal(err)
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.godepgraph.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVar(&path, "path", ".", "local path of packages")
	rootCmd.Flags().BoolVar(&web, "web", false, "serve a web server and show the depgraph in the webpage")
	rootCmd.Flags().IntVar(&port, "port", 7788, "port of web server, default 7788")
	rootCmd.Flags().IntVar(&depth, "depth", 5, "max depth, default 5")
}
