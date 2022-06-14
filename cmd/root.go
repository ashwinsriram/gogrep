/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"github.com/fatih/color"
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"
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
		grepSearch(args[1], args[0])
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
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

//search for a string in a file and return the line number and line with the string highlighted
func grepSearch(file string, search string) {
	//open the file
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)
	ln := 0
	firstMatch := true

	for fileScanner.Scan() {
		ln++
		if strings.Contains(fileScanner.Text(), search) {
			if firstMatch {
				color.Blue(file)
				firstMatch = false
			}
			fmt.Printf("%v: ",color.GreenString(strconv.Itoa(ln)))
			line := strings.Split(strings.TrimSpace(fileScanner.Text()), search)
			numParts := len(line) - 1
			for idx,part := range line {
				fmt.Print(part)
				if idx < numParts {
					color.New(color.FgRed).Print(search)
				}
			}
			fmt.Println()
		}
	}
}



