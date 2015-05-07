package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFsck(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("fsck")
	cmd.Output = "Git LFS fsck OK"

	testFileContent := "test data"
	h := sha256.New()
	io.WriteString(h, testFileContent)
	wantOid := hex.EncodeToString(h.Sum(nil))

	cmd.Before(func() {
		path := filepath.Join(repo.Path, ".git", "info", "attributes")
		repo.WriteFile(path, "*.dat filter=lfs -crlf\n")

		// Add a Git LFS object
		repo.WriteFile(filepath.Join(repo.Path, "a.dat"), testFileContent)
		repo.GitCmd("add", "a.dat")
		repo.GitCmd("commit", "-m", "a")
	})

	cmd.After(func() {
		// Verify test file exists as LFS object
		lfsObjectPath := filepath.Join(repo.Path, ".git", "lfs", "objects", wantOid[0:2], wantOid[2:4], wantOid)
		if _, err := os.Stat(lfsObjectPath); err != nil {
			t.Fatal(err)
		}

		// Corrupt the LFS object and verify that fsck detects corruption
		repo.WriteFile(lfsObjectPath, testFileContent+"CORRUPTION")
		err := doFsck(filepath.Join(repo.Path, ".git"))
		wantErr := &fsckError{"a.dat", wantOid}
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}