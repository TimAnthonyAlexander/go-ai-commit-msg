package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev" // will be set by goreleaser
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gh-smart-commit",
	Short: "AI-powered Git assistant using local Ollama",
	Long: `gh-smart-commit is a CLI tool that uses local Ollama models to generate
commit messages, code suggestions, and other Git-related assistance.

All operations are performed locally for privacy and work offline once
the Ollama model is downloaded.`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/gh-smart-commit.yaml)")
	rootCmd.PersistentFlags().String("ollama-host", "127.0.0.1:11434", "Ollama server host:port")
	rootCmd.PersistentFlags().String("model", "llama3:8b", "Ollama model to use")
	rootCmd.PersistentFlags().Float32("temperature", 0.3, "Model temperature (0.0-1.0)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")

	// Bind flags to viper
	viper.BindPFlag("ollama.host", rootCmd.PersistentFlags().Lookup("ollama-host"))
	viper.BindPFlag("ollama.model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("ollama.temperature", rootCmd.PersistentFlags().Lookup("temperature"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
			os.Exit(1)
		}

		// Search config in ~/.config/gh-smart-commit.yaml
		viper.AddConfigPath(home + "/.config")
		viper.SetConfigName("gh-smart-commit")
		viper.SetConfigType("yaml")
	}

	// Environment variables
	viper.SetEnvPrefix("GH_SMART_COMMIT")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err == nil && viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
} 