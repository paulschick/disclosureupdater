package config

import (
	"github.com/paulschick/disclosureupdater/common/constants"
	"github.com/spf13/viper"
)

// ConfKeys is a struct that holds the configuration keys
// for the main folders of the application
type ConfKeys struct {
	profile string
	keyMap  []*ConfKeyMap
}

// ConfKeyMap is a struct that holds the configuration key
// for a single folder of the application
type ConfKeyMap struct {
	contextKey string
	configKey  string
	defaultVal string
	currentVal string
}

// GetContextKey returns the urfave CLI context key
func (c *ConfKeyMap) GetContextKey() string {
	return c.contextKey
}

// GetConfigKey returns the viper configuration key
func (c *ConfKeyMap) GetConfigKey() string {
	return c.configKey
}

// GetDefaultVal returns the default value for the folder
func (c *ConfKeyMap) GetDefaultVal() string {
	return c.defaultVal
}

// GetCurrentVal returns the current value for the folder
func (c *ConfKeyMap) GetCurrentVal() string {
	return c.currentVal
}

// SetViperValue sets the current value for the folder to Viper
func (c *ConfKeyMap) SetViperValue(v *viper.Viper) {
	v.Set(c.configKey, c.currentVal)
}

// BuildConfKeyMaps builds the configuration keys for the main folders
func BuildConfKeyMaps(profile string) *ConfKeys {
	configKeys := map[string]string{
		"disclosuresFolder": constants.DefaultDisclosuresFolder,
		"imagesFolder":      constants.DefaultImageFolder,
		"ocrFolder":         constants.DefaultOcrFolder,
		"csvFolder":         constants.DefaultCsvFolder,
		"s3Folder":          constants.DefaultS3Folder,
	}

	confKeys := make([]*ConfKeyMap, len(configKeys))

	i := 0
	for k, v := range configKeys {
		confKeys[i] = &ConfKeyMap{
			contextKey: v,
			configKey:  profile + ".folders." + k,
			defaultVal: v,
			currentVal: v,
		}
		i++
	}

	return &ConfKeys{
		profile: profile,
		keyMap:  confKeys,
	}
}

// GetKeyMap returns the configuration key for a single folder based on the key used
// to set the value in the CLI
func (c *ConfKeys) GetKeyMap(ctxKey string) *ConfKeyMap {
	for _, v := range c.keyMap {
		if v.GetContextKey() == ctxKey {
			return v
		}
	}
	return nil
}

// GetFolders returns a list of folders to be created
func (c *ConfKeys) GetFolders() []string {
	folders := make([]string, len(c.keyMap))
	for i, v := range c.keyMap {
		folders[i] = v.GetCurrentVal()
	}
	return folders
}

// GetCommonDirs creates a new CommonDirs object for use in commands
func (c *ConfKeys) GetCommonDirs() *CommonDirs {
	return &CommonDirs{
		BaseFolder:        GetBaseFolder(),
		DataFolder:        GetDataFolder(),
		S3Folder:          c.GetKeyMap("s3").GetCurrentVal(),
		DisclosuresFolder: c.GetKeyMap("disclosures").GetCurrentVal(),
		ImageFolder:       c.GetKeyMap("images").GetCurrentVal(),
		OcrFolder:         c.GetKeyMap("ocr").GetCurrentVal(),
		CsvFolder:         c.GetKeyMap("csv").GetCurrentVal(),
	}
}
