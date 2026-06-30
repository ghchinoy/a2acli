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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghchinoy/a2acli/internal/oauth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	envName string

	// Flag vars for config env add
	addServiceURL string
	addTransport  string
	addToken      string
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
	envTransport := viper.GetString(envPrefix + "transport")

	// 3. Override global variables if they were NOT set by explicitly passed CLI flags.

	if !rootCmd.Flag("service-url").Changed && envURL != "" {
		serviceURL = envURL
	}
	if !rootCmd.Flag("token").Changed && envToken != "" {
		authToken = envToken
	}
	if !rootCmd.Flag("transport").Changed && envTransport != "" {
		transport = envTransport
	}
}

// defaultConfigPath returns the default XDG-compliant config file path.
func defaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	dir := filepath.Join(configDir, "a2acli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// saveConfig writes the current viper configuration back to disk.
func saveConfig() error {
	err := viper.WriteConfig()
	if err != nil {
		// If no config file is loaded (first time), WriteConfig fails.
		// Locate the default path and write there.
		path, err := defaultConfigPath()
		if err != nil {
			return err
		}
		verboseLog("no config file loaded; writing to default path: %s", path)
		return viper.WriteConfigAs(path)
	}
	return nil
}

// setupConfigCmd builds the `config` command group.
func setupConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:     "config",
		GroupID: GroupSystem,
		Short:   "View and manage client configuration",
		Long: `View the active configuration settings or manage named environment profiles.

Settings are loaded from the default configuration file ($HOME/.config/a2acli/config.yaml) 
and can be overridden by environment variables and command-line flags.`,
		Example: `  a2acli config
  a2acli config --env production
  a2acli config env list`,
		Run: runConfig,
	}

	// env group
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage named environment profiles",
		Long:  `Add, remove, list, and select named environments in the config.yaml.`,
	}

	// env add
	addCmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add or update a named environment",
		Long:  `Add a new named environment profile to config.yaml, or update an existing one.`,
		Example: `  a2acli config env add staging --service-url https://staging.example.com
  a2acli config env add prod -u https://prod.example.com --transport grpc
  a2acli config env add dev -u http://127.0.0.1:9001 --token my-static-token`,
		Args: cobra.ExactArgs(1),
		Run:  runConfigEnvAdd,
	}
	addCmd.Flags().StringVarP(&addServiceURL, "service-url", "u", "", "Base URL of the A2A service (required)")
	_ = addCmd.MarkFlagRequired("service-url")
	addCmd.Flags().StringVar(&addTransport, "transport", "", "Force transport: grpc, jsonrpc, rest")
	addCmd.Flags().StringVar(&addToken, "token", "", "Static auth token")

	// env remove
	removeCmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a named environment",
		Example: `  a2acli config env remove staging`,
		Args:    cobra.ExactArgs(1),
		Run:     runConfigEnvRemove,
	}

	// env use
	useCmd := &cobra.Command{
		Use:     "use <name>",
		Short:   "Set the default environment",
		Example: `  a2acli config env use prod`,
		Args:    cobra.ExactArgs(1),
		Run:     runConfigEnvUse,
	}

	// env list
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all named environments",
		Example: `  a2acli config env list
  a2acli config env list --output json`,
		Run: runConfigEnvList,
	}

	envCmd.AddCommand(addCmd, removeCmd, useCmd, listCmd)
	configCmd.AddCommand(envCmd)
	return configCmd
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
	if transport != "" {
		fmt.Printf("Transport: %s\n", transport)
	}
}

func runConfigEnvAdd(_ *cobra.Command, args []string) {
	name := args[0]
	prefix := fmt.Sprintf("envs.%s.", name)

	viper.Set(prefix+"service_url", addServiceURL)
	if addTransport != "" {
		switch strings.ToLower(addTransport) {
		case "grpc", "jsonrpc", "rest":
			viper.Set(prefix+"transport", strings.ToLower(addTransport))
		default:
			fatalf("invalid transport", fmt.Errorf("%q", addTransport), "Must be grpc, jsonrpc, or rest")
		}
	}
	if addToken != "" {
		viper.Set(prefix+"token", addToken)
	}

	if err := saveConfig(); err != nil {
		fatalf("failed to save config", err, "")
	}

	fmt.Printf("Environment %s added/updated.\n", name)
	fmt.Printf("  Service URL: %s\n", addServiceURL)
	if addTransport != "" {
		fmt.Printf("  Transport:   %s\n", addTransport)
	}
	fmt.Printf("\nUse it with: a2acli <command> --env %s\n", name)
}

func runConfigEnvRemove(_ *cobra.Command, args []string) {
	name := args[0]

	// Check if this is the default_env.
	if viper.GetString("default_env") == name {
		fmt.Printf("Warning: %s is currently set as your default_env.\n", name)
		viper.Set("default_env", "default")
	}

	envs := viper.GetStringMap("envs")
	if _, ok := envs[name]; !ok {
		fatalf("environment not found", fmt.Errorf("%q", name), "Run 'a2acli config env list' to see available environments")
	}

	delete(envs, name)
	viper.Set("envs", envs)

	if err := saveConfig(); err != nil {
		fatalf("failed to save config", err, "")
	}
	fmt.Printf("Environment %s removed.\n", name)
}

func runConfigEnvUse(_ *cobra.Command, args []string) {
	name := args[0]

	// Verify it exists first
	envs := viper.GetStringMap("envs")
	if _, ok := envs[name]; !ok && name != "default" {
		fatalf("environment not found", fmt.Errorf("%q", name), "Create it first with 'a2acli config env add'")
	}

	viper.Set("default_env", name)
	if err := saveConfig(); err != nil {
		fatalf("failed to save config", err, "")
	}
	fmt.Printf("Default environment set to %s.\n", name)
}

type jsonEnvOut struct {
	Name       string `json:"name"`
	ServiceURL string `json:"service_url"`
	Transport  string `json:"transport,omitempty"`
	HasToken   bool   `json:"has_token"`
	TokenState string `json:"token_state,omitempty"`
	IsDefault  bool   `json:"is_default"`
}

func runConfigEnvList(_ *cobra.Command, _ []string) {
	envs := viper.GetStringMap("envs")
	defaultEnv := viper.GetString("default_env")
	if defaultEnv == "" {
		defaultEnv = "default"
	}

	var list []jsonEnvOut
	for name := range envs {
		prefix := fmt.Sprintf("envs.%s.", name)
		urlVal := viper.GetString(prefix + "service_url")
		transVal := viper.GetString(prefix + "transport")
		tokenVal := viper.GetString(prefix + "token")

		hasToken := false
		tokenState := "none"
		if tokenVal != "" {
			hasToken = true
			tokenState = "static"
		} else {
			// Check token store
			if stored, err := oauth.LoadToken(urlVal); err == nil && stored != nil {
				hasToken = true
				if stored.IsExpired() {
					tokenState = "expired"
				} else {
					tokenState = "valid"
				}
			}
		}

		list = append(list, jsonEnvOut{
			Name:       name,
			ServiceURL: urlVal,
			Transport:  transVal,
			HasToken:   hasToken,
			TokenState: tokenState,
			IsDefault:  name == defaultEnv,
		})
	}

	// Always include "default" if not explicitly in the map
	if _, ok := envs["default"]; !ok {
		urlVal := "http://127.0.0.1:9001"
		hasToken := false
		tokenState := "none"
		if stored, err := oauth.LoadToken(urlVal); err == nil && stored != nil {
			hasToken = true
			if stored.IsExpired() {
				tokenState = "expired"
			} else {
				tokenState = "valid"
			}
		}
		list = append(list, jsonEnvOut{
			Name:       "default",
			ServiceURL: urlVal,
			HasToken:   hasToken,
			TokenState: tokenState,
			IsDefault:  defaultEnv == "default",
		})
	}

	if disableTUI {
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Printf("\nConfigured Environments:\n\n")
	for _, env := range list {
		prefix := "  - "
		if env.IsDefault {
			prefix = "  * " // mark default with star
		}
		fmt.Printf("%s%s  url=%s", prefix, env.Name, env.ServiceURL)
		if env.Transport != "" {
			fmt.Printf("  transport=%s", env.Transport)
		}
		if env.HasToken {
			fmt.Printf("  token=%s", env.TokenState)
		}
		if env.IsDefault {
			fmt.Printf("  (default)")
		}
		fmt.Println()
	}
	fmt.Println()
}
