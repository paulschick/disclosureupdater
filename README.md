# House of Representatives - Disclosure Download CLI

## Overview

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This CLI program is designed to download PDF financial disclosure reports for the
members of the House of Representatives. This tool downloads PDF reports, can upload the
PDFs to S3 storage, and can convert the PDFs to PNG or JPG image files. The purpose
of the image conversion is to be used with an OCR program to extract text and tables.

## Features

- **Initialize Environment**: Set up necessary directories and configuration files.
- **Configure S3 Settings**: Update Amazon S3 configuration for storing the PDFs.
- **Download Disclosure URLs**: Retrieve the latest list of disclosure URLs via XML files.
- **Download PDFs**: Download the PDF transaction reports.
- **Update Bucket Items**: Maintain an updated list of items in the S3 bucket.
- **Upload PDFs to S3**: Upload new PDFs to Amazon S3, ensuring the latest reports are stored.
- **Convert PDFs to Images**: Convert downloaded PDFs to PNG or JPG formats.
- **Cleanup Images**: Remove empty or failed image directories.

## Installation

To use Disclosure Download CLI, you must have Go installed on your machine. 
You can download and install Go from [here](https://go.dev/doc/install).

Once Go is installed, you can install the Disclosure Download CLI by running:

```shell
go get github.com/paulschick/disclosureupdater
```

## Usage

### Initialize

To initialize the environment for the first time, use:

```shell
disclosurecli initialize
```

### Configure S3 Settings

To configure Amazon S3 settings, you can either use the CLI or place the credentials in the ~/.aws directory. 
To use the CLI for configuration, run:

```shell
disclosurecli configure --s3-bucket [bucket_name] --s3-region [region] --s3-hostname [hostname] --s3-api-key [api_key] --s3-secret-key [secret_key]
```

If you prefer to use the `~/.aws` credentials file, ensure that you have the AWS SDK configured to load these 
credentials by setting the environment variable:

```shell
export AWS_SDK_LOAD_CONFIG=true
```

Then, create a credentials file at `~/.aws/credentials` with the following format:

```text
[default]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY
```

Replace `YOUR_ACCESS_KEY` and `YOUR_SECRET_KEY` with your actual AWS credentials.

### Download Disclosure URLs

To download the latest disclosure URLs, run:

```shell
disclosurecli update-urls
```

### Download PDFs

To download the transaction report PDFs, use:

```shell
disclosurecli download-pdfs
```

### Upload to S3

To upload the PDFs to S3, use:

```shell
disclosurecli upload-s3
```

If you download more PDF files after you've uploaded to S3 initially, you'll want to update the S3 index locally
to ensure you're only uploading files that aren't currently in S3:

```shell
disclosurecli update-bucket-items
```

### Convert PDFs to Images

To convert the PDFs to images, use:

```shell
disclosurecli convert-pdfs
# For JPG conversion
disclosurecli convert-pdfs --jpg
```

### Cleanup Images

To remove empty directories and failed image conversions, use:

```shell
disclosurecli cleanup-images
```

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## Additional Documentation

![CLI Usage](./docs/resources/CLI%20Diagram%20v1.png)

**Note**: This README provides a basic overview of the project. Ensure that all configurations, especially those 
related to AWS S3 authentication and environment settings, are accurately followed for the tool to function correctly. 
More detailed documentation may be necessary, especially for contributing guidelines and license details, 
depending on the project's complexity and additional requirements.
