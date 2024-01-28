package config

import (
	"github.com/paulschick/disclosureupdater/common/methods"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"path"
)

// ConfigurationBuilder is used to build the configuration object on application startup.
// This is responsible for creating the initial configuration directories and files,
// then building the configuration object from the context and any existing configuration.
type ConfigurationBuilder struct {
	v        *viper.Viper
	profile  string
	confKeys *ConfKeys
}

// NewConfigurationBuilder creates a new configuration builder used to configure
// the CommonDirs object for use within various commands
func NewConfigurationBuilder(profile string, ctx *cli.Context) (*ConfigurationBuilder, error) {
	v, err := InitializeViper()
	if err != nil {
		return nil, err
	}
	builder := &ConfigurationBuilder{
		v:        v,
		profile:  profile,
		confKeys: BuildConfKeyMaps(profile),
	}
	for _, confKeyMap := range builder.confKeys.keyMap {
		value := builder.extractValue(ctx, confKeyMap)
		confKeyMap.currentVal = value
	}
	return builder, nil
}

// extractValue extracts the value from the context, then the configuration, then the default value
// Used to build the full configuration object from the context and any existing configuration
func (c *ConfigurationBuilder) extractValue(ctx *cli.Context, confKeyMap *ConfKeyMap) string {
	value := ctx.String(confKeyMap.GetContextKey())
	if value == "" {
		value = c.v.GetString(confKeyMap.GetConfigKey())
		if value == "" {
			value = confKeyMap.GetDefaultVal()
		}
	} else {
		value = path.Join(GetDataFolder(), value)
	}
	return value
}

// build builds the configuration object and writes the Viper configuration file
func (c *ConfigurationBuilder) build() error {
	for _, confKeyMap := range c.confKeys.keyMap {
		confKeyMap.SetViperValue(c.v)
	}
	return c.v.WriteConfig()
}

// createDirectories creates the directories specified in the configuration
// if they don't already exist
func (c *ConfigurationBuilder) createDirectories() error {
	for _, dir := range c.confKeys.GetFolders() {
		if err := methods.TryCreateDirectories(dir); err != nil {
			return err
		}
	}
	return nil
}

// Build Creates directories and builds the configuration object
func (c *ConfigurationBuilder) Build() error {
	err := c.createDirectories()
	if err != nil {
		return err
	}
	err = c.build()
	if err != nil {
		return err
	}
	return nil
}

// GetCommonDirs wraps ConfKeys GetCommonDirs method
func (c *ConfigurationBuilder) GetCommonDirs() *CommonDirs {
	return c.confKeys.GetCommonDirs()
}
