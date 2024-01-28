package config

import (
	"os"
	"path"
)

// GetBaseFolder returns the base folder for the application which contains
// the config file and all data
func GetBaseFolder() string {
	return path.Join(os.Getenv("HOME"), ".disclosurecli")
}

// GetDataFolder returns the data folder for the application which contains
func GetDataFolder() string {
	return path.Join(GetBaseFolder(), "data")
}

// GetConfigFileName Currently just returns default
// allow this to be configurable
func GetConfigFileName() string {
	return "config.yaml"
}

// GetConfigProfile Currently just returns default
func GetConfigProfile() string {
	return "default"
}
