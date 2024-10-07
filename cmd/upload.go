package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/spf13/cobra"
)



var uploadCmd = &cobra.Command{
    Use:   "upload",
    Short: "Sync directory to S3 by comparing checksums",
    RunE: func(cmd *cobra.Command, args []string) error {
        return uploadToS3(inputDir, bucketName)
    },
}

func init() {
    rootCmd.AddCommand(uploadCmd)
    uploadCmd.Flags().StringVarP(&inputDir, "input", "i", "", "Input directory path")
    uploadCmd.Flags().StringVarP(&bucketName, "bucket", "b", "", "S3 bucket name")
    uploadCmd.Flags().StringVarP(&endpointURL, "endpoint", "e", "", "AWS Endpoint URL (for testing with LocalStack)")
    uploadCmd.Flags().StringVarP(&region, "region", "r", "us-east-1", "AWS Region")
    uploadCmd.Flags().BoolVarP(&deleteExtra, "delete", "d", false, "Delete files in S3 that are not present in the local directory")
    uploadCmd.MarkFlagRequired("input")
    uploadCmd.MarkFlagRequired("bucket")
}

func uploadToS3(inputDir, bucketName string) error {
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
        // Try to create the bucket
        _, err = s3Client.CreateBucket(&s3.CreateBucketInput{
            Bucket: aws.String(bucketName),
        })
        if err != nil {
            return fmt.Errorf("failed to create bucket: %v", err)
        }
    }

    // Map to store local files and their checksums
    localFiles := make(map[string]string)

    // Walk through the local directory and compute checksums
    err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !info.IsDir() {
            relativePath, err := filepath.Rel(inputDir, path)
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

    // Upload new or updated files
    for relativePath, localChecksum := range localFiles {
        s3Checksum, exists := s3Objects[relativePath]
        if !exists || localChecksum != s3Checksum {
            // File is new or has changed, upload it
            err := uploadFile(s3Client, bucketName, inputDir, relativePath)
            if err != nil {
                return err
            }
            fmt.Printf("Uploaded %s to s3://%s/%s\n", relativePath, bucketName, relativePath)
        } else {
            fmt.Printf("Skipped (unchanged): %s\n", relativePath)
        }
    }

    // Optionally delete files in S3 that are not in local directory
    if deleteExtra {
        for s3Key := range s3Objects {
            if _, exists := localFiles[s3Key]; !exists {
                // Delete the object from S3
                _, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
                    Bucket: aws.String(bucketName),
                    Key:    aws.String(s3Key),
                })
                if err != nil {
                    return fmt.Errorf("failed to delete object %s: %v", s3Key, err)
                }
                fmt.Printf("Deleted s3://%s/%s\n", bucketName, s3Key)
            }
        }
    }

    return nil
}

func uploadFile(s3Client *s3.S3, bucketName, inputDir, relativePath string) error {
    filePath := filepath.Join(inputDir, relativePath)
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = s3Client.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(relativePath),
        Body:   file,
    })
    return err
}
