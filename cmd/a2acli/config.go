// Copyright 2026 Google LLC
//
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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	envName string
)

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find config directory per XDG spec
		configDir, err := os.UserConfigDir()
		if err != nil {
			// Fallback for systems without UserConfigDir
			homeDir, err := os.UserHomeDir()
			if err == nil {
				configDir = filepath.Join(homeDir, ".config")
			}
		}

		if configDir != "" {
			a2aConfigDir := filepath.Join(configDir, "a2acli")
			viper.AddConfigPath(a2aConfigDir)

			// On some OS (like macOS), os.UserConfigDir returns ~/Library/Application Support.
			// Let's also explicitly add the standard XDG ~/.config/a2acli path for universality.
			homeDir, err := os.UserHomeDir()
			if err == nil {
				viper.AddConfigPath(filepath.Join(homeDir, ".config", "a2acli"))

				// Also support legacy ~/.a2acli.yaml if it exists
				viper.AddConfigPath(homeDir)
			}

			viper.SetConfigType("yaml")
			viper.SetConfigName("config")
		}
	}

	viper.SetEnvPrefix("A2ACLI")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in silently.
	_ = viper.ReadInConfig()

	// 1. Determine which environment to use
	targetEnv := envName
	if targetEnv == "" {
		// Fallback to default_env in config, or "default"
		targetEnv = viper.GetString("default_env")
		if targetEnv == "" {
			targetEnv = "default"
		}
	}

	// 2. Fetch the values for that specific environment
	envPrefix := fmt.Sprintf("envs.%s.", targetEnv)
	envURL := viper.GetString(envPrefix + "service_url")
	envToken := viper.GetString(envPrefix + "token")

	// 3. Override global variables if they were NOT set by explicitly passed CLI flags
	// (Cobra/Viper integration handles this cleanly if we ask Viper, but since we are binding
	// directly to vars in root.PersistentFlags, we manually check if the user left them as defaults).

	// If the user didn't explicitly pass a URL via the command line (-u), use the config URL
	if !rootCmd.Flag("service-url").Changed && envURL != "" {
		serviceURL = envURL
	}

	// If the user didn't explicitly pass a token (-t), use the config token
	if !rootCmd.Flag("token").Changed && envToken != "" {
		authToken = envToken
	}
}
func runConfig(_ *cobra.Command, _ []string) {
	fmt.Printf("Config File Used: %s\n", viper.ConfigFileUsed())

	targetEnv := envName
	if targetEnv == "" {
		targetEnv = viper.GetString("default_env")
		if targetEnv == "" {
			targetEnv = "default"
		}
	}
	fmt.Printf("Active Environment: %s\n", targetEnv)
	fmt.Printf("Service URL: %s\n", serviceURL)

	tokenStr := "<none>"
	if authToken != "" {
		tokenStr = "<set>"
	}
	fmt.Printf("Auth Token: %s\n", tokenStr)
}
