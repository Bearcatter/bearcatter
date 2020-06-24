package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bearcatter",
	Short: "bearcatter is a program that lets you interact with your Uniden scanner.",
	Long:  `Bearcatter is a system for controlling, recording and managing the Uniden SDS100 and SDS200.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bearcatter.yaml)")

	loggerLvl := rootCmd.PersistentFlags().String("log.level", "InfoLevel", "The level of log to show [default = InfoLevel]. Available options are (PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel)")

	switch *loggerLvl {
	case "TraceLevel":
		log.SetLevel(log.TraceLevel)
	case "DebugLevel":
		log.SetLevel(log.DebugLevel)
	case "InfoLevel":
		log.SetLevel(log.InfoLevel)
	case "WarnLevel":
		log.SetLevel(log.WarnLevel)
	case "ErrorLevel":
		log.SetLevel(log.ErrorLevel)
	case "FatalLevel":
		log.SetLevel(log.FatalLevel)
	case "PanicLevel":
		log.SetLevel(log.PanicLevel)
	default:
		log.Fatalf("Logrus logger level %s doesn't exist", *loggerLvl)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".bearcatter" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".bearcatter")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
