package cmd

import (
    "crypto/md5"
    "encoding/hex"
    "io"
    "os"
)

var (
    inputDir       string
	outputDir    string
    bucketName     string
    endpointURL    string
    region         string
    deleteExtra    bool
)

func computeMD5Checksum(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    hash := md5.New()
    if _, err := io.Copy(hash, file); err != nil {
        return "", err
    }

    checksum := hex.EncodeToString(hash.Sum(nil))
    return checksum, nil
}
