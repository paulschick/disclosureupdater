package main

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path"
)

type CliFunc func(cCtx *cli.Context) error

// main
// https://github.com/spf13/viper
func main() {
	var err error
	commonDirs := getCommonDirs()

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
					return initialize(commonDirs)
				},
			},
			{
				Name:      "configure",
				Aliases:   []string{"conf"},
				Usage:     "Initialize environment configuration",
				UsageText: "Write S3 configuration values to file",
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("Updating S3 configuration\n")
					s3Profile := config.S3ProfileFromCtx(cCtx)
					err := config.UpdateS3Config(s3Profile, commonDirs)
					if err != nil {
						return err
					}
					fmt.Printf("S3 configuration updated\n")
					return nil
				},
				Flags: []cli.Flag{
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
			{
				Name:  "update-urls",
				Usage: "Update the list of disclosure URLs",
				UsageText: "Update the list of disclosure URLs\n" +
					"   disclosurecli update-urls\n",
				Action: func(cCtx *cli.Context) error {
					return DownloadUrlsCmd(commonDirs)(cCtx)
				},
			},
			{
				Name:  "download-pdfs",
				Usage: "Download the PDFs",
				UsageText: "Download the PDFs\n" +
					"   disclosurecli download-pdfs\n",
				Action: func(cCtx *cli.Context) error {
					return DownloadPdfsCmd(commonDirs)(cCtx)
				},
			},
			{
				Name:  "update-bucket-items",
				Usage: "Update the list of bucket items",
				UsageText: "Update the list of bucket items\n" +
					"   disclosurecli update-bucket-items\n",
				Action: func(cCtx *cli.Context) error {
					return UpdateBucketItemIndex(commonDirs)(cCtx)
				},
			},
			{
				Name:      "upload-s3",
				Usage:     "Upload PDFs to S3 that are not present",
				UsageText: "disclosurecli upload-s3\n",
				Action: func(cCtx *cli.Context) error {
					return UploadPdfs(commonDirs)(cCtx)
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "update-index",
						Aliases: []string{"u"},
						Usage:   "Update the list of bucket items",
					},
				},
			},
			{
				Name:  "convert-pdfs",
				Usage: "Convert PDFs to JPGs",
				UsageText: "Convert PDFs to JPGs\n" +
					"   disclosurecli convert-pdfs\n",
				Action: func(cCtx *cli.Context) error {
					return PdfToJpg(commonDirs)(cCtx)
				},
			},
			{
				Name:  "ocr-jpgs",
				Usage: "OCR JPGs using gosseract",
				UsageText: "OCR JPGs using gosseract\n" +
					"   disclosurecli ocr-jpgs\n",
				Action: func(cCtx *cli.Context) error {
					return JpgOcr(commonDirs)(cCtx)
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

func getBaseDir() string {
	return path.Join(os.Getenv("HOME"), ".disclosurecli")
}

func getCommonDirs() *config.CommonDirs {
	return config.NewCommonDirs(getBaseDir())
}

func initialize(commonDirs *config.CommonDirs) error {
	var err error
	profile := "default"
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
