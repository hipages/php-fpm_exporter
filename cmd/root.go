// Copyright © 2018 Enrico Stahn <enrico.stahn@gmail.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmd contains the CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/hipages/php-fpm_exporter/phpfpm"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var log = logrus.New()

// Version that is being reported by the CLI
var Version string

var cfgFile, logLevel string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "php-fpm_exporter",
	Short: "Exports php-fpm metrics for prometheus",
	Long:  `php-fpm_exporter exports prometheus compatible metrics from php-fpm.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.php-fpm_exporter.yaml)")
	RootCmd.PersistentFlags().StringVar(&logLevel, "log.level", "info", "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]")
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

		// Search config in home directory with name ".php-fpm_exporter" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".php-fpm_exporter")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// initLogger configures the log level
func initLogger() {
	phpfpm.SetLogger(log)

	if value := os.Getenv("PHP_FPM_LOG_LEVEL"); value != "" {
		logLevel = value
	}

	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		lvl = logrus.InfoLevel
		log.Fatalf("Could not set log level to '%v'.", logLevel)
	}

	log.SetLevel(lvl)
}

func mapEnvVars(envs map[string]string, cmd *cobra.Command) {
	for env, flag := range envs {
		flag := cmd.Flags().Lookup(flag)
		flag.Usage = fmt.Sprintf("%v [env %v]", flag.Usage, env)
		if value := os.Getenv(env); value != "" {
			if err := flag.Value.Set(value); err != nil {
				log.Error(err)
			}
		}
	}
}
