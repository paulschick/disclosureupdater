package main

import (
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

// main
// TODO - implement read/write of configuration files using Viper
// https://github.com/spf13/viper
// Viper is already installed.
func main() {
	var err error

	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// TODO create .overrides
	// This will be auto-set by the program if the user changes any of the defaults,
	// like the name or location of the configuration file. This should be in a configuration directory
	// based on the operating system, like AppData or $HOME/.config/disclosurecli or $HOME/.disclosurecli
	// I think $HOME/.disclourecli would be best since it can be used on all operating systems.

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
							log.Printf("Data folder: %s\n", dataFolder)
							log.Printf("S3 bucket: %s\n", s3Bucket)
							log.Printf("S3 region: %s\n", s3Region)
							log.Printf("S3 hostname: %s\n", s3Hostname)

							viper.Set("profile", "default")
							viper.Set("dataFolder", dataFolder)
							viper.Set("s3Bucket", s3Bucket)
							viper.Set("s3Region", s3Region)
							viper.Set("s3Hostname", s3Hostname)
							return viper.WriteConfig()
						},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "data-folder",
								Aliases: []string{"d"},
								Usage:   "Data folder",
							},
							&cli.StringFlag{
								Name:    "s3-bucket",
								Aliases: []string{"b"},
								Usage:   "S3 bucket",
							},
							&cli.StringFlag{
								Name:    "s3-region",
								Aliases: []string{"r"},
								Usage:   "S3 region",
							},
							&cli.StringFlag{
								Name:    "s3-hostname",
								Aliases: []string{"e"},
								Usage:   "S3 hostname",
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
