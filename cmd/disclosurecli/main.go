package main

import (
	"github.com/paulschick/disclosureupdater/util"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path"
)

// main
// https://github.com/spf13/viper
func main() {
	var err error

	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// TODO - this is going to be the default configuration file and app directory
	overridesDir := path.Join(os.Getenv("HOME"), ".disclosurecli")

	err = util.TryCreateDirectories(overridesDir)
	if err != nil {
		log.Fatalf("Error creating overrides directory: %s", err)
	}
	overridesPath := path.Join(overridesDir, "overrides.yaml")

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}
	app := &cli.App{
		Name:                   "Disclosure Download CLI",
		Version:                "0.0.1",
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			{
				Name:    "initialize",
				Aliases: []string{"init"},
				Usage:   "Initialize environment",
				UsageText: "Create the necessary directories and configuration files. " +
					"Use this command to create the initial environment.",
				Action: func(cCtx *cli.Context) error {
					log.Println("Initializing environment")
					return initialize()
				},
			},
			{
				Name:    "configure",
				Aliases: []string{"conf"},
				Usage:   "Initialize environment configuration",
				UsageText: "Either load from an environment file, or write a new environment file " +
					"from passed command line flags",
				Subcommands: []*cli.Command{
					{
						Name:    "load",
						Aliases: []string{"l"},
						Usage:   "Load environment configuration from file",
						Action: func(cCtx *cli.Context) error {
							log.Println("Loading environment configuration from file")
							return nil
						},
					},
					{
						Name:    "write",
						Aliases: []string{"w"},
						Usage:   "Write environment configuration to file",
						Action: func(cCtx *cli.Context) error {
							log.Println("Writing environment configuration to file")
							dataFolder := cCtx.String("data-folder")
							s3Bucket := cCtx.String("s3-bucket")
							s3Region := cCtx.String("s3-region")
							s3Hostname := cCtx.String("s3-hostname")
							s3ApiKey := cCtx.String("s3-api-key")
							s3SecretKey := cCtx.String("s3-secret-key")
							log.Printf("Data folder: %s\n", dataFolder)
							log.Printf("S3 bucket: %s\n", s3Bucket)
							log.Printf("S3 region: %s\n", s3Region)
							log.Printf("S3 hostname: %s\n", s3Hostname)
							log.Printf("S3 API key: %s\n", s3ApiKey)
							log.Printf("S3 secret key: %s\n", s3SecretKey)

							//viper.Set("profile", "default")
							viper.Set("default.dataFolder", dataFolder)
							viper.Set("default.s3.s3Bucket", s3Bucket)
							viper.Set("default.s3.s3Region", s3Region)
							viper.Set("default.s3.s3Hostname", s3Hostname)
							viper.Set("default.s3.s3ApiKey", s3ApiKey)
							viper.Set("default.s3.s3SecretKey", s3SecretKey)
							log.Printf("Overrides path: %s\n", overridesPath)
							return viper.WriteConfig()
						},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "data-folder",
								Aliases: []string{"d"},
								Usage:   "Data folder",
								EnvVars: []string{"DATA_FOLDER"},
							},
							&cli.StringFlag{
								Name:    "s3-bucket",
								Aliases: []string{"b"},
								Usage:   "S3 bucket",
								EnvVars: []string{"S3_BUCKET"},
							},
							&cli.StringFlag{
								Name:    "s3-region",
								Aliases: []string{"r"},
								Usage:   "S3 region",
								EnvVars: []string{"S3_REGION"},
							},
							&cli.StringFlag{
								Name:    "s3-hostname",
								Aliases: []string{"e"},
								Usage:   "S3 hostname",
								EnvVars: []string{"S3_HOSTNAME"},
							},
							&cli.StringFlag{
								Name:    "s3-api-key",
								Aliases: []string{"k"},
								Usage:   "S3 API key",
								EnvVars: []string{"S3_API_KEY"},
							},
							&cli.StringFlag{
								Name:    "s3-secret-key",
								Aliases: []string{"s"},
								Usage:   "S3 secret key",
								EnvVars: []string{"S3_SECRET_KEY"},
							},
						},
						Subcommands: []*cli.Command{
							{
								Name:    "override",
								Aliases: []string{"o"},
								Usage:   "Override configuration file location and name",
								UsageText: "Create a new file path and/or file name for " +
									"the configuration file",
								Action: func(cCtx *cli.Context) error {
									log.Println("Overriding configuration file location and name")
									newPath := cCtx.String("config-directory")
									log.Printf("New path: %s\n", newPath)
									newFile := cCtx.String("config-file")
									log.Printf("New file: %s\n", newFile)

									if newPath != "" || newFile != "" {
										log.Println("Overriding configuration file location and name")
										// Create/Update overrides file
										viper.SetConfigFile(overridesPath)
										viper.SetConfigType("yaml")
										viper.Set("configDirectory", newPath)
										viper.Set("configFile", newFile)
										err = viper.WriteConfig()
										if err != nil {
											log.Fatalf("Error writing overrides file: %s", err)
										}

										// Then write the new configuration file
										viper.SetConfigFile(path.Join(newPath, newFile))
										viper.SetConfigType("yaml")
										err = viper.WriteConfig()
										if err != nil {
											log.Fatalf("Error writing new configuration file: %s", err)
										}
									}
									return nil
								},
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:    "config-directory",
										Usage:   "Configuration file path - Provide an existing directory",
										Aliases: []string{"d"},
									},
									&cli.StringFlag{
										Name: "config-file",
										Usage: "Configuration file name - Provide a new file name with " +
											".yaml extension.\nMust provide a yaml configuration file.",
										Aliases: []string{"f"},
									},
								},
							},
						},
					},
				},
				Action: func(cCtx *cli.Context) error {
					log.Println("Initializing data set")
					return nil
				},
			},
		},
	}

	err = runApp(app)
	if err != nil {
		log.Fatalf("Error running app: %s", err)
	}
}

func runApp(app *cli.App) error {
	return app.Run(os.Args)
}

type Directories struct {
	BaseDir string
	DataDir string
	S3Dir   string
}

func NewDirectories() *Directories {
	baseDir := path.Join(os.Getenv("HOME"), ".disclosurecli")
	dataDir := path.Join(baseDir, "data")
	s3Dir := path.Join(baseDir, "s3")
	return &Directories{
		BaseDir: baseDir,
		DataDir: dataDir,
		S3Dir:   s3Dir,
	}
}

func (d *Directories) CreateDirectories() error {
	err := util.TryCreateDirectories(d.DataDir)
	if err != nil {
		return err
	}
	err = util.TryCreateDirectories(d.S3Dir)
	if err != nil {
		return err
	}
	return nil
}

func (d *Directories) CreateExampleConfigFile() error {
	defaultProfile := NewDefaultProfile(d)
	configYamlFile := defaultProfile.YamlFile
	viper.SetConfigFile(configYamlFile)
	viper.SetConfigType("yaml")
	viper.Set("default.dataFolder", defaultProfile.DataFolder)
	viper.Set("default.s3.s3Bucket", defaultProfile.S3.S3Bucket)
	viper.Set("default.s3.s3Region", defaultProfile.S3.S3Region)
	viper.Set("default.s3.s3Hostname", defaultProfile.S3.S3Hostname)
	viper.Set("default.s3.s3ApiKey", defaultProfile.S3.S3ApiKey)
	viper.Set("default.s3.s3SecretKey", defaultProfile.S3.S3SecretKey)

	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

type S3Profile struct {
	S3Bucket    string
	S3Region    string
	S3Hostname  string
	S3ApiKey    string
	S3SecretKey string
}

type Profile struct {
	Name       string
	YamlFile   string
	DataFolder string
	S3         *S3Profile
}

func NewDefaultProfile(directories *Directories) *Profile {
	return &Profile{
		Name:       "default",
		YamlFile:   path.Join(directories.BaseDir, "config.yaml"),
		DataFolder: directories.DataDir,
		S3: &S3Profile{
			S3Bucket:    "<bucket>",
			S3Region:    "<region>",
			S3Hostname:  "<hostname>",
			S3ApiKey:    "<api-key>",
			S3SecretKey: "<secret-key>",
		},
	}
}

func initialize() error {
	var err error
	directories := NewDirectories()
	err = directories.CreateDirectories()
	if err != nil {
		return err
	}
	err = directories.CreateExampleConfigFile()
	if err != nil {
		return err
	}

	// Create example configuration files
	return nil
}
