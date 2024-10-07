package cmd

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
    Use:   "download",
    Short: "Sync directory with S3 bucket",
    RunE: func(cmd *cobra.Command, args []string) error {
        return downloadToLocal(bucketName, outputDir)
    },
}

func init() {
    rootCmd.AddCommand(syncCmd)
    syncCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Local directory path")
    syncCmd.Flags().StringVarP(&bucketName, "bucket", "b", "", "S3 bucket name")
    syncCmd.Flags().StringVarP(&endpointURL, "endpoint", "e", "", "AWS Endpoint URL (for testing with LocalStack)")
    syncCmd.Flags().StringVarP(&region, "region", "r", "us-east-1", "AWS Region")
    syncCmd.Flags().BoolVarP(&deleteExtra, "delete", "d", false, "Delete files that are not present in the source")
    syncCmd.MarkFlagRequired("input")
    syncCmd.MarkFlagRequired("bucket")
}

// Function to sync from S3 to local directory
func downloadToLocal(bucketName, localDir string) error {
    config := &aws.Config{
        Region: aws.String(region),
    }

    if endpointURL != "" {
        config.Endpoint = aws.String(endpointURL)
        config.S3ForcePathStyle = aws.Bool(true)
        // Set static credentials for LocalStack
        config.Credentials = credentials.NewStaticCredentials("test", "test", "")
    }

    sess, err := session.NewSession(config)
    if err != nil {
        return fmt.Errorf("failed to create AWS session: %v", err)
    }

    s3Client := s3.New(sess)

    // Ensure the bucket exists
    _, err = s3Client.HeadBucket(&s3.HeadBucketInput{
        Bucket: aws.String(bucketName),
    })
    if err != nil {
        return fmt.Errorf("failed to access bucket: %v", err)
    }

    // Map to store local files and their checksums
    localFiles := make(map[string]string)

    // Walk through the local directory and compute checksums
    err = filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !info.IsDir() {
            relativePath, err := filepath.Rel(localDir, path)
            if err != nil {
                return err
            }

            checksum, err := computeMD5Checksum(path)
            if err != nil {
                return err
            }

            localFiles[relativePath] = checksum
        }

        return nil
    })
    if err != nil {
        return err
    }

    // Map to store S3 objects and their ETags
    s3Objects := make(map[string]string)

    // List objects in the S3 bucket
    err = s3Client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
        Bucket: aws.String(bucketName),
    }, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
        for _, obj := range page.Contents {
            key := *obj.Key
            etag := strings.Trim(*obj.ETag, "\"") // Remove quotes from ETag
            s3Objects[key] = etag
        }
        return !lastPage
    })
    if err != nil {
        return fmt.Errorf("failed to list objects in bucket: %v", err)
    }

    // Download new or updated files
    for s3Key, s3Checksum := range s3Objects {
        localChecksum, exists := localFiles[s3Key]
        if !exists || localChecksum != s3Checksum {
            // File is new or has changed, download it
            err := downloadFile(s3Client, bucketName, localDir, s3Key)
            if err != nil {
                return err
            }
            fmt.Printf("Downloaded s3://%s/%s to %s\n", bucketName, s3Key, s3Key)
        } else {
            fmt.Printf("Skipped (unchanged): %s\n", s3Key)
        }
    }

    // Optionally delete local files that are not in S3
    if deleteExtra {
        for localKey := range localFiles {
            if _, exists := s3Objects[localKey]; !exists {
                // Delete the local file
                localFilePath := filepath.Join(localDir, localKey)
                err := os.Remove(localFilePath)
                if err != nil {
                    return fmt.Errorf("failed to delete local file %s: %v", localFilePath, err)
                }
                fmt.Printf("Deleted local file: %s\n", localKey)
            }
        }
    }

    return nil
}

// Function to download a file from S3
func downloadFile(s3Client *s3.S3, bucketName, localDir, s3Key string) error {
    output, err := s3Client.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(s3Key),
    })
    if err != nil {
        return err
    }
    defer output.Body.Close()

    // Create the local file path
    localFilePath := filepath.Join(localDir, s3Key)
    localDirPath := filepath.Dir(localFilePath)

    // Ensure the directory exists
    err = os.MkdirAll(localDirPath, os.ModePerm)
    if err != nil {
        return err
    }

    // Create the local file
    localFile, err := os.Create(localFilePath)
    if err != nil {
        return err
    }
    defer localFile.Close()

    // Copy the content
    _, err = io.Copy(localFile, output.Body)
    return err
}