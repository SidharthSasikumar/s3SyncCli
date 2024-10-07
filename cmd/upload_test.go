package cmd

import (
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "testing"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestuploadToS3(t *testing.T) {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "localstack/localstack",
        ExposedPorts: []string{"4566/tcp"},
        Env: map[string]string{
            "SERVICES": "s3",
        },
        WaitingFor: wait.ForListeningPort("4566/tcp"),
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

    // Set the endpointURL for the sync function
    endpointURL = endpoint
    region = "us-east-1"

    // Create a temporary directory with some files
    tempDir, err := ioutil.TempDir("", "testsync")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // Create initial test files and directories
    os.Mkdir(filepath.Join(tempDir, "mvs"), 0755)
    ioutil.WriteFile(filepath.Join(tempDir, "index.json"), []byte(`{"key": "value"}`), 0644)
    ioutil.WriteFile(filepath.Join(tempDir, "mvs", "indexmvs.json"), []byte(`{"mvskey": "mvsvalue"}`), 0644)

    // Set the inputDir and bucketName
    inputDir = tempDir
    bucketName = "test-bucket"
    deleteExtra = false

    // Run the sync function (initial upload)
    err = uploadToS3(inputDir, bucketName)
    if err != nil {
        t.Fatalf("uploadToS3 failed: %v", err)
    }

    // Verify that the files were uploaded
    sess, err := session.NewSession(&aws.Config{
        Region:           aws.String(region),
        Endpoint:         aws.String(endpointURL),
        Credentials:      credentials.NewStaticCredentials("test", "test", ""),
        S3ForcePathStyle: aws.Bool(true),
    })
    if err != nil {
        t.Fatalf("Failed to create AWS session: %v", err)
    }

    s3Client := s3.New(sess)

    // List objects in the bucket after initial upload
    initialObjects, err := listS3Objects(s3Client, bucketName)
    if err != nil {
        t.Fatalf("Failed to list objects after initial upload: %v", err)
    }

    expectedKeys := map[string]bool{
        "index.json":        true,
        "mvs/indexmvs.json": true,
    }

    if !compareKeys(expectedKeys, initialObjects) {
        t.Errorf("Initial upload: Expected keys %v, got %v", expectedKeys, initialObjects)
    }

    // Modify one file and add a new file
    ioutil.WriteFile(filepath.Join(tempDir, "index.json"), []byte(`{"key": "new value"}`), 0644)
    ioutil.WriteFile(filepath.Join(tempDir, "newfile.json"), []byte(`{"new": "file"}`), 0644)

    // Remove one file locally
    os.Remove(filepath.Join(tempDir, "mvs", "indexmvs.json"))

    // Run the sync function again without deletion
    err = uploadToS3(inputDir, bucketName)
    if err != nil {
        t.Fatalf("uploadToS3 failed: %v", err)
    }

    // List objects in the bucket after second sync
    secondObjects, err := listS3Objects(s3Client, bucketName)
    if err != nil {
        t.Fatalf("Failed to list objects after second sync: %v", err)
    }

    expectedKeysAfterSecondSync := map[string]bool{
        "index.json":        true,
        "mvs/indexmvs.json": true, // Should still exist since we didn't delete extra files
        "newfile.json":      true,
    }

    if !compareKeys(expectedKeysAfterSecondSync, secondObjects) {
        t.Errorf("Second sync without deletion: Expected keys %v, got %v", expectedKeysAfterSecondSync, secondObjects)
    }

    // Now run sync with deletion of extra files
    deleteExtra = true
    err = uploadToS3(inputDir, bucketName)
    if err != nil {
        t.Fatalf("uploadToS3 with deleteExtra failed: %v", err)
    }

    // List objects in the bucket after deletion
    finalObjects, err := listS3Objects(s3Client, bucketName)
    if err != nil {
        t.Fatalf("Failed to list objects after deletion: %v", err)
    }

    expectedKeysAfterDeletion := map[string]bool{
        "index.json":   true,
        "newfile.json": true,
    }

    if !compareKeys(expectedKeysAfterDeletion, finalObjects) {
        t.Errorf("Final sync with deletion: Expected keys %v, got %v", expectedKeysAfterDeletion, finalObjects)
    }
}

func listS3Objects(s3Client *s3.S3, bucketName string) (map[string]bool, error) {
    objectKeys := make(map[string]bool)

    err := s3Client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
        Bucket: aws.String(bucketName),
    }, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
        for _, obj := range page.Contents {
            objectKeys[*obj.Key] = true
        }
        return !lastPage
    })
    return objectKeys, err
}

func compareKeys(expected, actual map[string]bool) bool {
    if len(expected) != len(actual) {
        return false
    }
    for key := range expected {
        if !actual[key] {
            return false
        }
    }
    return true
}
