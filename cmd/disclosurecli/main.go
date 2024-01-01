package main

import (
	"github.com/paulschick/disclosureupdater/config"
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

func initialize() error {
	var err error
	profile := "default"
	baseDir := path.Join(os.Getenv("HOME"), ".disclosurecli")
	commonDirs := config.NewCommonDirs(baseDir)
	err = commonDirs.CreateDirectories()
	if err != nil {
		return err
	}
	s3Default := config.S3DefaultProfile{
		S3Bucket:   "<bucket>",
		S3Region:   "<region>",
		S3Hostname: "<hostname>",
	}
	v := config.BuildViper(profile, commonDirs, &s3Default)
	err = v.WriteConfig()
	if err != nil {
		return err
	}

	// Create example configuration files
	return nil
}
