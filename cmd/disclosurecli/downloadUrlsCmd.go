package main

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/downloader"
	"github.com/paulschick/disclosureupdater/util"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func DownloadUrlsCmd(commonDirs *config.CommonDirs) CliFunc {
	var err error

	return func(cCtx *cli.Context) error {
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
		return err
	}
}
