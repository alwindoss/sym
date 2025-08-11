// Package cmd defines all the cli commands of the 'sym' application
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alwindoss/sym/internal/sym"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sym",
	Short: "A fast symlink farm manager for dotfiles and packages",
	Long: `Sym is a modern symlink farm manager written in Go, designed to help you 
organize and deploy your dotfiles, configuration files, and software packages.

By creating symbolic links from a centralized source directory to target locations,
Sym allows you to maintain a clean, version-controlled collection of your 
configuration files while making them appear in their expected system locations.

Key features:
  • Sym packages by creating symlinks to target directories  
  • Unsym packages by safely removing managed symlinks
  • Resym packages for easy updates and reorganization
  • Dry-run mode to preview changes before applying them
  • Verbose output for detailed operation logging
  • Safe conflict detection and resolution

Perfect for managing dotfiles, development environments, and system configurations
across multiple machines with version control integration.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	RunE: func(cmd *cobra.Command, args []string) error {
		config := &sym.Config{
			SymDir:    symDir,
			TargetDir: targetDir,
			Verbose:   verbose,
			Simulate:  simulate,
			Delete:    deleteFlag,
			ReSym:     resym,
		}

		config.Packages = args
		if len(config.Packages) == 0 {
			fmt.Fprintf(os.Stderr, "error: No packages specified\n\n\n")
			cmd.Help()
			os.Exit(1)
		}

		// Convert to absolute paths
		var err error
		config.SymDir, err = filepath.Abs(config.SymDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving stow directory: %v\n", err)
			os.Exit(1)
		}

		config.TargetDir, err = filepath.Abs(config.TargetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving target directory: %v\n", err)
			os.Exit(1)
		}

		if config.Verbose {
			fmt.Printf("Stow dir: %s\n", config.SymDir)
			fmt.Printf("Target dir: %s\n", config.TargetDir)
		}

		for _, pkg := range config.Packages {
			if err := sym.ProcessPackage(config, pkg); err != nil {
				err = fmt.Errorf("error processing package '%s': %w", pkg, err)
				return err
			}
		}
		return nil
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

var (
	symDir, targetDir                             string
	verbose, simulate, deleteFlag, resym, version bool
)

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sym.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Sym flags
	rootCmd.Flags().StringVarP(&symDir, "dir", "d", ".", "sym directory")
	rootCmd.Flags().StringVarP(&targetDir, "target", "t", "..", "target directory")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().BoolVarP(&simulate, "simulate", "n", false, "simulate actions (dry run)")
	rootCmd.Flags().BoolVarP(&deleteFlag, "delete", "D", false, "delete/unsym packages")
	rootCmd.Flags().BoolVarP(&resym, "resym", "R", false, "resym packages (unsym then sym)")
	rootCmd.Flags().BoolVar(&version, "version", false, "show version")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".sym" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".sym")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
