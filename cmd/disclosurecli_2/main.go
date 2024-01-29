package main

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/commands"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"os"
)

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}
	app := &cli.App{
		Name:                   "Disclosure Download CLI",
		Version:                "0.0.2",
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			{
				Name:  "test-commander",
				Usage: "test new impl",
				Action: func(cc *cli.Context) error {
					cmd := commands.NewCommander(cc)
					defer cmd.Cleanup()
					profile := cmd.GetProfile()
					commonDirs := cmd.GetCommonDirs()
					log := cmd.GetLogger()
					log.Info("test-commander", zap.String("profile", profile),
						zap.Any("commonDirs", commonDirs))
					fmt.Printf("%s\n", commonDirs.ToString())
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "profile",
						Aliases: []string{"p"},
						Usage:   "profile to use for configuration",
					},
					&cli.StringFlag{
						Name:    "images",
						Aliases: []string{"i"},
						Usage:   "folder to store images within the data folder",
					},
					&cli.StringFlag{
						Name:    "disclosures",
						Aliases: []string{"d"},
						Usage:   "folder to store disclosures within the data folder",
					},
					&cli.StringFlag{
						Name:    "ocr",
						Aliases: []string{"o"},
						Usage:   "folder to store ocr files within the data folder",
					},
					&cli.StringFlag{
						Name:    "csv",
						Aliases: []string{"c"},
						Usage:   "folder to store csv files within the data folder",
					},
					&cli.StringFlag{
						Name:    "s3",
						Aliases: []string{"s"},
						Usage:   "folder to store s3 files within the data folder",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error running app: %v\n", err)
	}
}
