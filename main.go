package main

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/paulschick/disclosureupdater/s3client"
	"log"
)

func main() {
	fmt.Println("House of Representatives Data Updater")
	configuration := model.Configure()
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
	s3Service, err := s3client.NewS3Service(configuration)
	if err != nil {
		log.Fatalf("Error creating S3 service: %s", err)
	}
	err = s3Service.WriteBucketObjects()
	if err != nil {
		log.Fatalf("Error writing bucket objects: %s", err)
	} else {
		fmt.Println("Successfully wrote bucket objects")
	}
	// TODO end S3 service testing
}
