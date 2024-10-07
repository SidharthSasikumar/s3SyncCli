package cmd

import (
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestDownloadToLocal(t *testing.T) {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "localstack/localstack",
        ExposedPorts: []string{"4566/tcp"},
        Env: map[string]string{
            "SERVICES": "s3",
        },
        WaitingFor: wait.ForLog("Ready."),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("Failed to start container: %v", err)
    }
    defer func() {
        _ = container.Terminate(ctx)
    }()

    host, err := container.Host(ctx)
    if err != nil {
        t.Fatalf("Failed to get container host: %v", err)
    }

    port, err := container.MappedPort(ctx, "4566")
    if err != nil {
        t.Fatalf("Failed to get mapped port: %v", err)
    }

    endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

    // Set the endpointURL and region
    endpointURL = endpoint
    region = "us-east-1"

    // Set the bucket name
    bucketName = "test-bucket"

    // Initialize AWS session
    config := &aws.Config{
        Region:           aws.String(region),
        Endpoint:         aws.String(endpointURL),
        Credentials:      credentials.NewStaticCredentials("test", "test", ""),
        S3ForcePathStyle: aws.Bool(true),
    }
    sess, err := session.NewSession(config)
    if err != nil {
        t.Fatalf("Failed to create AWS session: %v", err)
    }
    s3Client := s3.New(sess)

    // Create the bucket
    _, err = s3Client.CreateBucket(&s3.CreateBucketInput{
        Bucket: aws.String(bucketName),
    })
    if err != nil {
        t.Fatalf("Failed to create bucket: %v", err)
    }

    // Upload test files to S3
    testFiles := map[string]string{
        "index.json":        `{"key": "value"}`,
        "mvs/indexmvs.json": `{"mvskey": "mvsvalue"}`,
    }
    for key, content := range testFiles {
        _, err = s3Client.PutObject(&s3.PutObjectInput{
            Bucket: aws.String(bucketName),
            Key:    aws.String(key),
            Body:   strings.NewReader(content),
        })
        if err != nil {
            t.Fatalf("Failed to upload object %s: %v", key, err)
        }
    }

    // Create a temporary local directory
    tempDir, err := ioutil.TempDir("", "testdownload")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // Set the outputDir
    outputDir = tempDir

    // Run the download function
    err = downloadToLocal(bucketName, outputDir)
    if err != nil {
        t.Fatalf("downloadToLocal failed: %v", err)
    }

    // Verify that the files were downloaded
    for key, content := range testFiles {
        localFilePath := filepath.Join(outputDir, key)
        data, err := ioutil.ReadFile(localFilePath)
        if err != nil {
            t.Fatalf("Failed to read local file %s: %v", localFilePath, err)
        }
        if string(data) != content {
            t.Errorf("Content mismatch for %s: expected %s, got %s", key, content, string(data))
        }
    }

    // Modify a file in S3 and add a new file
    testFiles["index.json"] = `{"key": "new value"}`
    testFiles["newfile.json"] = `{"new": "file"}`
    _, err = s3Client.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String("index.json"),
        Body:   strings.NewReader(testFiles["index.json"]),
    })
    if err != nil {
        t.Fatalf("Failed to upload modified object index.json: %v", err)
    }
    _, err = s3Client.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String("newfile.json"),
        Body:   strings.NewReader(testFiles["newfile.json"]),
    })
    if err != nil {
        t.Fatalf("Failed to upload new object newfile.json: %v", err)
    }

    // Remove one file from S3
    _, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String("mvs/indexmvs.json"),
    })
    if err != nil {
        t.Fatalf("Failed to delete object mvs/indexmvs.json: %v", err)
    }
    delete(testFiles, "mvs/indexmvs.json")

    // Run the download function again without deletion
    err = downloadToLocal(bucketName, outputDir)
    if err != nil {
        t.Fatalf("downloadToLocal failed: %v", err)
    }

    // Verify that the updated files are downloaded and deleted files are still present locally
    for key, content := range testFiles {
        localFilePath := filepath.Join(outputDir, key)
        data, err := ioutil.ReadFile(localFilePath)
        if err != nil {
            t.Fatalf("Failed to read local file %s: %v", localFilePath, err)
        }
        if string(data) != content {
            t.Errorf("Content mismatch for %s: expected %s, got %s", key, content, string(data))
        }
    }

    // Check that the deleted S3 file still exists locally
    deletedFilePath := filepath.Join(outputDir, "mvs/indexmvs.json")
    if _, err := os.Stat(deletedFilePath); os.IsNotExist(err) {
        t.Errorf("Deleted S3 object mvs/indexmvs.json should still exist locally")
    }

    // Now run download with deletion of extra local files
    deleteExtra = true
    err = downloadToLocal(bucketName, outputDir)
    if err != nil {
        t.Fatalf("downloadToLocal with deleteExtra failed: %v", err)
    }

    // Verify that the deleted S3 file is now deleted locally
    if _, err := os.Stat(deletedFilePath); !os.IsNotExist(err) {
        t.Errorf("Deleted S3 object mvs/indexmvs.json should have been deleted locally")
    }
}
