/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gogrep",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")
		hidden, _ := cmd.Flags().GetBool("hidden")
		binary, _ := cmd.Flags().GetBool("binary")
		ignoreErrors, _ := cmd.Flags().GetBool("ignore-errors")
		if recursive {
			recursiveSearch(args[0], args[1], hidden, binary, ignoreErrors)
		} else {
			grepSearch(args[0], args[1], binary)
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gogrep.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("recursive", "r", false, "Recursively search a directory")
	rootCmd.Flags().BoolP("hidden", ".", false, "Search hidden files")
	rootCmd.Flags().BoolP("binary", "b", false, "Allow for non utf8 characters")
	rootCmd.Flags().BoolP("ignore-errors", "i", false, "Ignore all errors")
}

func recursiveSearch(search string, dir string, hidden bool, binary bool, ignoreErrors bool) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if !ignoreErrors {
				log.Println("gogrep: ", err)
			}
			return filepath.SkipDir
		}
		if info.IsDir() && (filepath.Base(path)[0] == '.' && !hidden) {
			return filepath.SkipDir
		}
		if !info.IsDir() && (hidden || filepath.Base(path)[0] != '.') {
			grepSearch(search, path, binary)
		}
		return nil
	})
	if err != nil && !ignoreErrors {
		log.Println("gogrep: ", err)
	}
}

//search for a string in a file and return the line number and line with the string highlighted
func grepSearch(search string, file string, binary bool) {
	//open the file
	f, _ := os.Open(file)
	defer f.Close()
	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)
	ln := 0
	firstMatch := true

	for fileScanner.Scan() {
		ln++
		if strings.Contains(fileScanner.Text(), search) && (utf8.ValidString(fileScanner.Text()) || binary) {
			if firstMatch {
				color.Blue(file)
				firstMatch = false
			}
			fmt.Printf("%v: ", color.GreenString(strconv.Itoa(ln)))
			line := strings.Split(strings.TrimSpace(fileScanner.Text()), search)
			numParts := len(line) - 1
			for idx, part := range line {
				fmt.Print(part)
				if idx < numParts {
					color.New(color.FgRed).Print(search)
				}
			}
			fmt.Println()
		}
	}
}
