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
	"sync"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var wgGrep sync.WaitGroup
var wgPrint sync.WaitGroup

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gogrep",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")
		hidden, _ := cmd.Flags().GetBool("hidden")
		binary, _ := cmd.Flags().GetBool("binary")
		ignoreErrors, _ := cmd.Flags().GetBool("ignore-errors")
		invert, _ := cmd.Flags().GetBool("invert")
		excludeExt, _ := cmd.Flags().GetStringSlice("exclude-ext")
		ext, _ := cmd.Flags().GetStringSlice("ext")
		excludeDir, _ := cmd.Flags().GetStringSlice("exclude-dir")
		excludeExtMap := make(map[string]bool)
		for _, exclude := range excludeExt {
			excludeExtMap[exclude] = true
		}
		extMap := make(map[string]bool)
		for _, e := range ext {
			extMap[e] = true
		}
		excludeDirMap := make(map[string]bool)
		for _, exclude := range excludeDir {
			excludeDirMap[exclude] = true
		}
		if recursive {
			recursiveSearch(args[0], args[1], hidden, binary, ignoreErrors, invert, excludeExtMap, extMap, excludeDirMap)
		} else {
			grepSearch(args[0], args[1], binary, invert)
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
	rootCmd.Flags().BoolP("invert", "v", false, "Returns all lines that do not match the pattern")
	rootCmd.Flags().StringSliceP("exclude-ext", "X", []string{}, "Exclude extensions from the search. Only works in recursive mode")
	rootCmd.Flags().StringSliceP("ext", "x", []string{}, "Only include certain extensions. Only works in recursive mode")
	rootCmd.Flags().StringSliceP("exclude-dir", "D", []string{}, "Exclude directories from the search. Only works in recursive mode")
}

func recursiveSearch(search string, dir string, hidden bool, binary bool, ignoreErrors bool, invert bool, excludeExtMap map[string]bool, extMap map[string]bool, excludeDirMap map[string]bool) {
	resChan := make(chan string)
	guard := make(chan struct{}, 128)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if !ignoreErrors {
				log.Println("gogrep: ", err)
			}
			return filepath.SkipDir
		}
		if info.IsDir() && (filepath.Base(path)[0] == '.' && !hidden) && filepath.Base(path) != "." {
			return filepath.SkipDir
		}
		if info.IsDir() && excludeDirMap[filepath.Base(path)] {
			return filepath.SkipDir
		}
		if len(extMap) > 0 && extMap[filepath.Ext(path)] {
			return nil
		}
		if excludeExtMap[filepath.Ext(path)] {
			return nil
		}

		if !info.IsDir() && (hidden || filepath.Base(path)[0] != '.') {
			wgGrep.Add(1)
			wgPrint.Add(1)
			guard <- struct{}{}
			go recursiveGrep(search, path, binary, resChan, guard, invert)
			go recursivePrint(path, resChan)
		}
		return nil
	})
	if err != nil && !ignoreErrors {
		log.Println("gogrep: ", err)
	}
	wgGrep.Wait()
	wgPrint.Wait()
	close(resChan)
}

//search for a string in a file and return the line number and line with the string highlighted
func grepSearch(search string, file string, binary bool, invert bool) {
	//open the file
	f, _ := os.Open(file)
	defer f.Close()
	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)
	ln := 0

	for fileScanner.Scan() {
		ln++
		if (strings.Contains(fileScanner.Text(), search) == !invert && len(fileScanner.Text()) > 0) && (utf8.ValidString(fileScanner.Text()) || binary) {
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

func recursiveGrep(search string, file string, binary bool, resChan chan string, guard chan struct{}, invert bool) {
	defer wgGrep.Done()
	defer func() { <-guard }()
	//open the file
	f, _ := os.Open(file)
	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)
	ln := 0
	res := color.GreenString(file) + "\n"
	for fileScanner.Scan() {
		ln++
		if (strings.Contains(fileScanner.Text(), search) == !invert && len(fileScanner.Text()) > 0) && (utf8.ValidString(fileScanner.Text()) || binary) {
			res += fmt.Sprintf("%s: ", color.GreenString(strconv.Itoa(ln)))
			line := strings.Split(strings.TrimSpace(fileScanner.Text()), search)
			numParts := len(line) - 1
			for idx, part := range line {
				res += fmt.Sprint(part)
				if idx < numParts {
					res += color.RedString(search)
				}
			}
			res += "\n"
		}
	}
	f.Close()
	if len(strings.Split(res, "\n")) == 2 {
		res = ""
	} else {
		res += "\n"
	}
	resChan <- res
}

func recursivePrint(path string, resChan chan string) {
	defer wgPrint.Done()
	res := <-resChan
	if res != "" {
		os.Stdout.Write([]byte(res))
	}
}
