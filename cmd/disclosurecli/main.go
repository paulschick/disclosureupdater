package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

type CliConfig struct {
	DataFolder  string
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3Secret    string
	S3Endpoint  string
}

func NewCliConfig(ctx *cli.Context) *CliConfig {
	return &CliConfig{
		DataFolder:  ctx.String("data-folder"),
		S3Bucket:    ctx.String("s3-bucket"),
		S3Region:    ctx.String("s3-region"),
		S3AccessKey: ctx.String("s3-access-key"),
		S3Secret:    ctx.String("s3-secret-key"),
		S3Endpoint:  ctx.String("s3-endpoint"),
	}
}

func printFlags(ctx *cli.Context) error {
	cliConfig := NewCliConfig(ctx)
	log.Printf("Data folder: %s\n", cliConfig.DataFolder)
	log.Printf("S3 bucket: %s\n", cliConfig.S3Bucket)
	log.Printf("S3 region: %s\n", cliConfig.S3Region)
	log.Printf("S3 access key: %s\n", cliConfig.S3AccessKey)
	log.Printf("S3 secret key: %s\n", cliConfig.S3Secret)
	log.Printf("S3 endpoint: %s\n", cliConfig.S3Endpoint)
	return nil
}

func main() {
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
				Aliases: []string{"i"},
				Usage:   "Initialize the full data set",
				Action: func(cCtx *cli.Context) error {
					log.Println("Initializing data set")
					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "Update the data set",
				Action: func(cCtx *cli.Context) error {
					log.Println("Updating data set")
					return nil
				},
			},
			{
				Name:  "s3",
				Usage: "Manage S3",
				UsageText: "Manage S3\n" +
					"   disclosurecli s3 refresh\n" +
					"   disclosurecli -d ./data s3 refresh\n" +
					"   disclosurecli -d ./data s3 -b bucketname -r us-region" +
					" -a accesskey -s secretkey -e https://s3-endpoint.com refresh\n" +
					"   disclosurecli s3 print",
				Aliases: []string{"s"},
				Subcommands: []*cli.Command{
					{
						Name:    "refresh",
						Aliases: []string{"r"},
						Usage:   "Refresh list of S3 objects in the bucket",
						Action: func(cCtx *cli.Context) error {
							log.Println("Refreshing S3 objects")
							return printFlags(cCtx)
						},
					},
					{
						Name:  "print",
						Usage: "Print list of S3 objects in the bucket",
						Action: func(cCtx *cli.Context) error {
							log.Println("Printing S3 objects")
							return printFlags(cCtx)
						},
					},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "s3-bucket",
						Aliases: []string{"b"},
						EnvVars: []string{"S3_BUCKET"},
						Usage:   "S3 bucket name",
					},
					&cli.StringFlag{
						Name:    "s3-region",
						Aliases: []string{"r"},
						EnvVars: []string{"S3_REGION"},
						Usage:   "S3 region",
						Value:   "us-east-1",
					},
					&cli.StringFlag{
						Name:    "s3-access-key",
						EnvVars: []string{"S3_ACCESS_KEY"},
						Aliases: []string{"a"},
						Usage:   "S3 access key",
					},
					&cli.StringFlag{
						Name:    "s3-secret-key",
						Aliases: []string{"s"},
						EnvVars: []string{"S3_SECRET_KEY"},
						Usage:   "S3 secret key",
					},
					&cli.StringFlag{
						Name:    "s3-endpoint",
						Aliases: []string{"e"},
						EnvVars: []string{"S3_ENDPOINT"},
						Usage:   "S3 endpoint",
						Value:   "https://ewr1.vultrobjects.com",
					},
				},
			},
		},
		// TODO - be able to provide a .env file instead of these flags
		// Should be listed in help, and the naming needs to be consistent
		// example: data-folder -> DATA_FOLDER
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "use-env",
				Aliases:  []string{"u"},
				Usage:    "Use an environment variable file instead of flags",
				Category: "setup",
			},
			&cli.StringFlag{
				Name:     "env-file",
				Aliases:  []string{"f"},
				Usage:    "Environment variable file",
				Category: "setup",
				Value:    ".env",
			},
			&cli.StringFlag{
				Name:     "data-folder",
				Aliases:  []string{"d"},
				Usage:    "data folder, defaults to ./data",
				Value:    "./data",
				Category: "data",
			},
		},
		Action: func(ctx *cli.Context) error {
			dataFolder := ctx.String("data-folder")
			s3Bucket := ctx.String("s3-bucket")
			s3Region := ctx.String("s3-region")
			s3AccessKey := ctx.String("s3-access-key")
			s3Secret := ctx.String("s3-secret-key")
			s3Endpoint := ctx.String("s3-endpoint")
			log.Printf("Data folder: %s\n", dataFolder)
			log.Printf("S3 bucket: %s\n", s3Bucket)
			log.Printf("S3 region: %s\n", s3Region)
			log.Printf("S3 access key: %s\n", s3AccessKey)
			log.Printf("S3 secret key: %s\n", s3Secret)
			log.Printf("S3 endpoint: %s\n", s3Endpoint)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error running app: %s", err)
	}
}

//func main() {
//	fmt.Println("House of Representatives Data Updater")
//	configuration := model.Configure()
// TODO add to CLI program
//
//// Create directory if it does not exist, including subdirs
//err := util.TryCreateDirectories(configuration.DataFolder)
//if err != nil {
//	panic(err)
//}
//
//downloadUrls := downloader.GenerateAllZipUrls()
//disclosureDownloads := make([]*downloader.DisclosureDownload, len(downloadUrls))
//
//fmt.Println(downloadUrls)
//
//for i := 0; i < len(downloadUrls); i++ {
//	disclosureDownloads[i] = downloader.NewDisclosureDownload(downloadUrls[i], configuration.DataFolder)
//	fmt.Println(disclosureDownloads[i].ToString())
//}
//
//err = downloader.DownloadZipsIfNotPresent(disclosureDownloads)
//if err != nil {
//	panic(err)
//}
//
//err = util.CreatePdfDownloadDirectory(configuration.DataFolder)
//if err != nil {
//	panic(err)
//}
//var downloadMembers []*model.Member
//downloadMembers, err = downloader.GetTransactionReportMembers(disclosureDownloads, configuration.DataFolder)
//if err != nil {
//	panic(err)
//}
//fmt.Printf("Downloading %d PDFs\n", len(downloadMembers))
//
//downloadables := make([]*downloader.Downloadable, len(downloadMembers))
//for i, member := range downloadMembers {
//	fmt.Printf("Downloading %s\n", member.BuildPdfUrl())
//	downloadables[i] = &downloader.Downloadable{
//		Url:   member.BuildPdfUrl(),
//		Bytes: nil,
//		Fp:    member.BuildPdfFilePath(configuration.DataFolder),
//	}
//}
//downloadables, err = downloader.DownloadMultiple(downloadables)
//if err != nil {
//	log.Fatalf("Error downloading PDFs: %s", err)
//}
//for _, download := range downloadables {
//	err = os.WriteFile(download.Fp, download.Bytes, 0644)
//	if err != nil {
//		log.Fatalf("Error writing PDF: %s", err)
//	}
//}
// TODO end add to CLI program

// TODO deprecate this section
//err = S3(configuration)
//if err != nil {
//	log.Fatalf("Error uploading to S3: %s", err)
//} else {
//	fmt.Println("Successfully uploaded to S3")
//}
// TODO end deprecation section

// TODO S3 service testing
//	s3Service, err := s3client.NewS3Service(configuration)
//	if err != nil {
//		log.Fatalf("Error creating S3 service: %s", err)
//	}
//	err = s3Service.WriteBucketObjects()
//	if err != nil {
//		log.Fatalf("Error writing bucket objects: %s", err)
//	} else {
//		fmt.Println("Successfully wrote bucket objects")
//	}
//	// TODO end S3 service testing
//}
