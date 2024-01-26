package config

import (
	"github.com/paulschick/disclosureupdater/logger"
	"github.com/paulschick/disclosureupdater/util"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"path"
)

type ConfigurationValue struct {
	Key   string
	Value string
}

type ConfigurationBuilder struct {
	v                 *viper.Viper
	profile           string
	dataFolder        ConfigurationValue
	disclosuresFolder ConfigurationValue
	imageFolder       ConfigurationValue
	ocrFolder         ConfigurationValue
	csvFolder         ConfigurationValue
	s3Folder          ConfigurationValue
}

func NewConfigurationBuilder(profile string, ctx *cli.Context) (*ConfigurationBuilder, error) {
	v, err := InitializeViper()
	if err != nil {
		return nil, err
	}
	builder := &ConfigurationBuilder{
		v:       v,
		profile: profile,
	}
	builder.initialize(ctx)
	return builder, nil
}

func (c *ConfigurationBuilder) extractValue(ctx *cli.Context, ctxKey, configKey, defaultValue string) string {
	value := ctx.String(ctxKey)
	if value == "" {
		value = c.v.GetString(configKey)
		if value == "" {
			value = defaultValue
		}
	}
	return value
}

func (c *ConfigurationBuilder) initialize(ctx *cli.Context) {
	dataFolderKey := c.profile + "dataFolder"
	dataFolderValue := GetDataFolder()
	c.dataFolder = ConfigurationValue{
		Key:   dataFolderKey,
		Value: dataFolderValue,
	}
	imageFolderKey := c.profile + "imageFolder"
	imageFolderValue := c.extractValue(ctx, "images", imageFolderKey,
		path.Join(dataFolderValue, DefaultImageFolder))
	c.imageFolder = ConfigurationValue{
		Key:   imageFolderKey,
		Value: imageFolderValue,
	}
	disclosuresFolderKey := c.profile + "disclosuresFolder"
	disclosuresFolderValue := c.extractValue(ctx, "disclosures", disclosuresFolderKey,
		path.Join(dataFolderValue, DefaultDisclosuresFolder))
	c.disclosuresFolder = ConfigurationValue{
		Key:   disclosuresFolderKey,
		Value: disclosuresFolderValue,
	}
	ocrFolderKey := c.profile + "ocrFolder"
	ocrFolderValue := c.extractValue(ctx, "ocr", ocrFolderKey,
		path.Join(dataFolderValue, DefaultOcrFolder))
	c.ocrFolder = ConfigurationValue{
		Key:   ocrFolderKey,
		Value: ocrFolderValue,
	}
	csvFolderKey := c.profile + "csvFolder"
	csvFolderValue := c.extractValue(ctx, "csv", csvFolderKey,
		path.Join(dataFolderValue, DefaultCsvFolder))
	c.csvFolder = ConfigurationValue{
		Key:   csvFolderKey,
		Value: csvFolderValue,
	}
	s3FolderKey := c.profile + "s3Folder"
	s3FolderValue := path.Join(dataFolderValue, DefaultS3Folder)
	c.s3Folder = ConfigurationValue{
		Key:   s3FolderKey,
		Value: s3FolderValue,
	}
}

func (c *ConfigurationBuilder) build() error {
	c.v.Set(c.dataFolder.Key, c.dataFolder.Value)
	c.v.Set(c.imageFolder.Key, c.imageFolder.Value)
	c.v.Set(c.disclosuresFolder.Key, c.disclosuresFolder.Value)
	c.v.Set(c.ocrFolder.Key, c.ocrFolder.Value)
	c.v.Set(c.csvFolder.Key, c.csvFolder.Value)
	c.v.Set(c.s3Folder.Key, c.s3Folder.Value)
	return c.v.WriteConfig()
}

func (c *ConfigurationBuilder) createDirectories() error {
	dirs := []string{
		c.dataFolder.Value,
		c.imageFolder.Value,
		c.disclosuresFolder.Value,
		c.ocrFolder.Value,
		c.csvFolder.Value,
		c.s3Folder.Value,
	}
	for _, dir := range dirs {
		if err := util.TryCreateDirectories(dir); err != nil {
			return err
		}
	}
	return nil
}

// InitConfig creates the base directory and creates the config file
func InitConfig() error {
	if err := util.TryCreateDirectories(GetBaseFolder()); err != nil {
		return err
	}
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path.Join(GetBaseFolder(), GetConfigFileName()))
	if err := v.WriteConfig(); err != nil {
		return err
	}
	return nil
}

// ConfigureDirectories is run after the initial configuration has been created
// this can be called to update the configuration with new values
// Call this to get the up-to-date configuration
func ConfigureDirectories(ctx *cli.Context) (*CommonDirs, error) {
	log := logger.Logger

	profile := ctx.String("profile")
	if profile == "" {
		profile = "default"
	}

	builder, err := NewConfigurationBuilder(profile, ctx)
	if err != nil {
		log.Error("Error initializing configuration builder", zap.Error(err))
		return nil, err
	}

	err = builder.createDirectories()
	if err != nil {
		log.Error("Error creating directories", zap.Error(err))
		return nil, err
	}

	err = builder.build()
	if err != nil {
		log.Error("Error writing configuration", zap.Error(err))
		return nil, err
	}

	return &CommonDirs{
		DataFolder:        builder.dataFolder.Value,
		ImageFolder:       builder.imageFolder.Value,
		DisclosuresFolder: builder.disclosuresFolder.Value,
		OcrFolder:         builder.ocrFolder.Value,
		CsvFolder:         builder.csvFolder.Value,
		S3Folder:          builder.s3Folder.Value,
	}, nil
}
