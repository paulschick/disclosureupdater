package main

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/downloader"
	"github.com/paulschick/disclosureupdater/s3client"
	"github.com/paulschick/disclosureupdater/util"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path"
	"strings"
)

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
					downloadUrls := downloader.GenerateAllZipUrls()
					disclosureDownloads := make([]*downloader.DisclosureDownload, len(downloadUrls))
					printStrs := make([]string, len(downloadUrls))
					for i, url := range downloadUrls {
						printStrs[i] = url + ",\n"
					}
					fmt.Printf("Updating disclosures for the following URLs:\n%s\n", printStrs)
					currentYear := util.CurrentYear()
					fmt.Printf("Updating for current year %d if present\n", currentYear)
					for i := 0; i < len(downloadUrls); i++ {
						disclosureDownloads[i] = downloader.NewDisclosureDownload(downloadUrls[i], commonDirs.DataFolder)
						urlWithYear := strings.Split(disclosureDownloads[i].Url, "FD.zip")[0]
						urlSplit := strings.Split(urlWithYear, "/")
						urlYear := urlSplit[len(urlSplit)-1]
						if urlYear == fmt.Sprintf("%d", currentYear) {
							fmt.Printf("URL with current year: %s\n", disclosureDownloads[i].Url)
							zipPath := disclosureDownloads[i].ZipPath
							xmlPath := disclosureDownloads[i].XmlPath
							// delete both if present
							if disclosureDownloads[i].ZipIsPresent() {
								fmt.Printf("Deleting zip file: %s\n", zipPath)
								err = os.Remove(zipPath)
								if err != nil {
									fmt.Printf("Error deleting zip file: %s\n", err)
									return err
								}
							}
							if disclosureDownloads[i].XmlIsPresent() {
								fmt.Printf("Deleting xml file: %s\n", xmlPath)
								err = os.Remove(xmlPath)
								if err != nil {
									fmt.Printf("Error deleting xml file: %s\n", err)
									return err
								}
							}
						}
						fmt.Println(disclosureDownloads[i].ToString())
					}
					err = downloader.DownloadZipsIfNotPresent(disclosureDownloads)
					if err != nil {
						fmt.Printf("Error downloading disclosures: %s\n", err)
						return err
					}
					return nil
				},
			},
			{
				Name:    "test-s3",
				Aliases: []string{"t"},
				Usage:   "Test S3 connection",
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("\n\tDISCLOSURE CLI\n")
					fmt.Printf("\n|          Test S3            |\n")
					fmt.Printf("------------------------------\n")
					fmt.Printf("Testing S3 connection\n")
					s3Profile := config.S3ProfileFromConfig(commonDirs, "default")
					fmt.Printf("S3 Bucket: %s\n", s3Profile.GetBucket())
					fmt.Printf("S3 Region: %s\n", s3Profile.GetRegion())
					fmt.Printf("S3 Hostname: %s\n", s3Profile.GetHostname())
					service, err := s3client.NewS3ServiceV2(s3Profile)
					if err != nil {
						fmt.Printf("Error creating S3 service: %s\n", err)
						return err
					}
					fmt.Printf("Initialized s3Service\n")
					err = service.CreateNewBucket()
					if err != nil {
						fmt.Printf("Error creating bucket: %s\n", err)
						return err
					}
					fmt.Printf("Created bucket if not exists operation complete\n")
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
