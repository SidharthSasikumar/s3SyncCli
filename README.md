S3 Uploader CLI Tool
====================

S3 Uploader is a Go-based Command Line Interface (CLI) tool that **syncs** files and directories to an AWS S3 bucket, maintaining the same directory hierarchy. It compares files using checksums and uploads only new or modified files. It's built using the Cobra framework and includes unit tests that utilize LocalStack to mock AWS services.

Table of Contents
-----------------

-   [Features](#features)
-   [Prerequisites](#prerequisites)
-   [Installation](#installation)
-   [Usage](#usage)
-   [Configuration](#configuration)
-   [Running Tests](#running-tests)
-   [Project Structure](#project-structure)
-   [Dependencies](#dependencies)
-   [Contributing](#contributing)
-   [License](#license)
-   [Additional Information](#additional-information)
-   [Contact](#contact)

* * * * *

Features
--------

-   **Syncs** files and directories to S3 while preserving the directory structure.
-   Compares files using **checksums** to upload only new or changed files.
-   Optionally **deletes** files in S3 that no longer exist locally.
-   Utilizes the Cobra framework for a robust CLI experience.
-   Supports testing with LocalStack to mock AWS S3 services.
-   Provides verbose output to track the sync process.

Prerequisites
-------------

-   **Go (Golang)**: Version 1.13 or higher is required. Download from the [official website](https://golang.org/dl/).
-   **AWS Account**: For uploading to real S3 buckets.
-   **Docker**: Required for running LocalStack during testing. Download from here.
-   **AWS CLI (Optional)**: For configuring AWS credentials and region.
    -   Install from the [official AWS CLI page](https://aws.amazon.com/cli/).

Installation
------------

1.  **Clone the Repository**

    ```bash
    git clone https://github.com/yourusername/s3uploader.git

    cd s3uploader
2.  **Initialize Go Modules**

    ```bash
    go mod tidy
3.  **Install Dependencies**

    ```bash
    go get ./...
    ```
4.  **Build the Application**

    ```bash
    go build -o s3uploader
    ```
Usage
-----

The CLI tool provides a command to sync directories to S3, comparing checksums to upload only new or changed files.

### Command Syntax

```bash
./s3uploader upload -i <input_directory> -b <bucket_name> [--region <aws_region>] [--endpoint <endpoint_url>] [--delete]
```

### Options

-   `-i, --input`: **(Required)** Path to the input directory you want to sync.
-   `-b, --bucket`: **(Required)** Name of the S3 bucket to sync with.
-   `-r, --region`: *(Optional)* AWS region where the bucket is located (default: `us-east-1`).
-   `-e, --endpoint`: *(Optional)* Custom AWS endpoint URL (useful for testing with LocalStack).
-   `-d, --delete`: *(Optional)* Delete files in S3 that are not present in the local directory.

### Examples

#### **Sync Local Directory with S3 Bucket**

```bash
./s3uploader upload -i /path/to/inputdirectory -b my-s3-bucket
```

#### **Sync with Deletion of Extra S3 Objects**

```bash
./s3uploader upload -i /path/to/inputdirectory -b my-s3-bucket --delete
```

#### **Sync Using LocalStack for Testing**

```bash
`./s3uploader upload -i /path/to/inputdirectory -b test-bucket -e http://localhost:4566`
```

* * * * *

Configuration
-------------

### AWS Credentials and Configuration

**You are expected to configure your AWS credentials and settings in the `~/.aws/config` and `~/.aws/credentials` files.**

The AWS SDK uses these files to authenticate and configure the AWS region.

#### **1\. AWS Credentials File**

Create or update the credentials file at `~/.aws/credentials`:

ini

Copy code

`[default]
aws_access_key_id = YOUR_ACCESS_KEY_ID
aws_secret_access_key = YOUR_SECRET_ACCESS_KEY`

#### **2\. AWS Config File**

Create or update the config file at `~/.aws/config`:

ini

Copy code

`[default]
region = YOUR_AWS_REGION`

**Note:** Replace `YOUR_ACCESS_KEY_ID`, `YOUR_SECRET_ACCESS_KEY`, and `YOUR_AWS_REGION` with your actual AWS credentials and preferred region.

#### **Alternative: Using Environment Variables**

You can also set AWS credentials and region using environment variables:

```bash
export AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_ACCESS_KEY
export AWS_DEFAULT_REGION=YOUR_AWS_REGION
```

### AWS CLI Configuration (Optional)

If you have the AWS CLI installed, you can configure your credentials using:

```bash
aws configure
```

This command will prompt you to enter your AWS Access Key ID, Secret Access Key, and default region.

* * * * *

Running Tests
-------------

Unit tests are provided to ensure the CLI works as expected. Tests utilize LocalStack to mock AWS services.

### Prerequisites for Testing

-   **Docker**: Ensure Docker is running on your system.

### Creating Test Files

The tests automatically create temporary directories and test files needed for testing. You do not need to manually create any test files. The `upload_test.go` file contains the logic to generate and manipulate test data during the tests.

### Running the Tests

```bash
go test ./cmd -v
```

### What the Tests Do

-   **Spin up** a LocalStack Docker container running S3.
-   **Create** a temporary directory with test files.
-   **Run** the CLI tool to sync files to the mocked S3 service.
-   **Modify** test files to simulate changes.
-   **Verify** that:
    -   New or modified files are uploaded.
    -   Unchanged files are not re-uploaded.
    -   Files deleted locally are deleted from S3 when the `--delete` flag is used.
-   **Tear down** the LocalStack container after tests complete.

### Understanding the Test File (`upload_test.go`)

The `upload_test.go` file contains tests that:

-   Verify the syncing functionality of the application.
-   Ensure that the application correctly uploads new or changed files.
-   Confirm that unchanged files are skipped.
-   Check that files deleted locally are also deleted from S3 when using the `--delete` flag.

**Key Points:**

-   The tests **automatically handle** the creation and deletion of test files and directories.
-   Tests use **dummy AWS credentials** since LocalStack doesn't require valid credentials.
-   The tests are designed to be **self-contained** and **repeatable**.

* * * * *

Project Structure
-----------------

```python
`s3uploader/
├── cmd/
│   ├── root.go          # Cobra root command
│   ├── upload.go        # Upload command implementation (sync functionality)
│   └── upload_test.go   # Unit tests using LocalStack
├── go.mod               # Go module file
├── go.sum               # Checksums for dependencies
├── main.go              # Main application entry point
└── README.md            # Documentation`

```
* * * * *

Dependencies
------------

-   **Cobra**: For building the CLI interface.
    -   [github.com/spf13/cobra](https://github.com/spf13/cobra)
-   **AWS SDK for Go**: For interacting with AWS services.
    -   [github.com/aws/aws-sdk-go](https://github.com/aws/aws-sdk-go)
-   **Testcontainers-Go**: For managing Docker containers in tests.
    -   [github.com/testcontainers/testcontainers-go](https://github.com/testcontainers/testcontainers-go)
-   **LocalStack**: Mock AWS services for testing.
    -   Docker Image: `localstack/localstack`

* * * * *

Contributing
------------

Contributions are welcome! Please follow these steps:

1.  **Fork the Repository**

    Click the "Fork" button at the top right of this page.

2.  **Clone Your Fork**

    ```bash
    git clone https://github.com/yourusername/s3uploader.git
    cd s3uploader
    ```

3.  **Create a Feature Branch**

    ```bash
    git checkout -b feature/your-feature-name
    ```

4.  **Make Changes and Commit**

    ```bash
    git commit -am 'Add new feature'
    ```

5.  **Push to Your Fork**

    ```bash
    git push origin feature/your-feature-name
    ```

6.  **Submit a Pull Request**

    Go to the original repository and click on "Pull Requests" to submit your changes for review.

* * * * *

License
-------

This project is licensed under the MIT License - see the LICENSE file for details.

* * * * *

Additional Information
----------------------

### Handling AWS Regions

The tool defaults to the `us-east-1` region. If you need to change the region, you can:

-   Use the `--region` flag when running the command.
-   Set the `AWS_REGION` environment variable.
-   Specify the region in your `~/.aws/config` file.

### Checksums and ETags

-   The tool uses **MD5 checksums** to compare local files with S3 objects.
-   **ETags** in S3 are used for comparison and are assumed to be the MD5 checksum of the object.
    -   Note: For multipart uploads, ETags are not simple MD5 checksums. This tool assumes files are uploaded in a single PUT operation.

### Error Handling

If you encounter errors during sync, the CLI will output the error message. Common issues include:

-   **Missing AWS Credentials**: Ensure your AWS credentials are configured as described in the [Configuration](#configuration) section.
-   **Incorrect AWS Region**: Verify that the region you're specifying matches the region where your bucket is located.
-   **Non-existent S3 Bucket**: Ensure the bucket name you provide exists or that you have permissions to create it.
-   **Network Connectivity Issues**: Check your internet connection and firewall settings.

### Extending the Tool

Feel free to extend the tool to suit your needs. Possible enhancements:

-   Add progress bars or more verbose logging.
-   Implement concurrency for faster uploads.
-   Add support for excluding certain files or directories.
-   Handle large files and multipart uploads correctly.
-   Add support for AWS session tokens or assume roles.

* * * * *

Contact
-------

For any questions or support, please open an issue on the GitHub repository or contact the maintainer at sidharth.sasikumar07@gmail.com.