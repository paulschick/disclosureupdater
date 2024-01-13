package main

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/downloader"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/urfave/cli/v2"
	"os"
)

func DownloadPdfsCmd(commonDirs *config.CommonDirs) CliFunc {
	var err error

	return func(c *cli.Context) error {
		fmt.Printf("Downloading PDFs\n")
		downloadUrls := downloader.GenerateAllZipUrls()
		disclosureDownloads := make([]*downloader.DisclosureDownload, len(downloadUrls))
		for i := 0; i < len(downloadUrls); i++ {
			disclosureDownloads[i] = downloader.NewDisclosureDownload(downloadUrls[i], commonDirs.DataFolder)
		}
		var downloadMembers []*model.Member
		downloadMembers, err = downloader.GetTransactionReportMembers(disclosureDownloads, commonDirs.DataFolder)
		if err != nil {
			fmt.Printf("Error getting transaction report members: %s\n", err)
			return err
		}
		fmt.Printf("Downloading %d PDFs\n", len(downloadMembers))
		downloadables := make([]*downloader.Downloadable, len(downloadMembers))
		for i, member := range downloadMembers {
			fmt.Printf("Downloading %s\n", member.BuildPdfUrl())
			downloadables[i] = &downloader.Downloadable{
				Url:   member.BuildPdfUrl(),
				Bytes: nil,
				Fp:    member.BuildPdfFilePath(commonDirs.DataFolder),
			}
		}
		downloadables, err = downloader.DownloadMultiple(downloadables)
		if err != nil {
			fmt.Printf("Error downloading PDFs: %s\n", err)
			return err
		}
		for _, download := range downloadables {
			err = os.WriteFile(download.Fp, download.Bytes, 0644)
			if err != nil {
				fmt.Printf("Error writing PDF: %s\n", err)
				return err
			}
		}
		return err
	}
}
