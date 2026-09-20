package cmd

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type objectStub struct {
	body  io.Reader
	calls int
}

func (s *objectStub) GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	s.calls++
	return &s3.GetObjectOutput{Body: io.NopCloser(s.body)}, nil
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("download interrupted") }
func TestSafeDownload(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	client := &objectStub{body: strings.NewReader("new")}
	for _, key := range []string{"../outside", "/absolute", "x/../../outside", `..\outside`} {
		if downloadFile(client, "test", root, key) == nil {
			t.Fatalf("accepted %q", key)
		}
	}
	if client.calls != 0 {
		t.Fatal("requested unsafe key")
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if downloadFile(client, "test", root, "link/file") == nil {
		t.Fatal("escaped root")
	}
	target := filepath.Join(root, "keep")
	os.WriteFile(target, []byte("old"), 0600)
	client.body = brokenReader{}
	if downloadFile(client, "test", root, "keep") == nil {
		t.Fatal("ignored failed download")
	}
	content, _ := os.ReadFile(target)
	if string(content) != "old" {
		t.Fatal("destroyed existing file")
	}
	client.body = strings.NewReader("complete")
	if err := downloadFile(client, "test", root, "nested/file"); err != nil {
		t.Fatal(err)
	}
	content, _ = os.ReadFile(filepath.Join(root, "nested/file"))
	if string(content) != "complete" {
		t.Fatal("incorrect download")
	}
}
