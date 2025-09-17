package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorage(t *testing.T) {
	dir := t.TempDir()
	storage, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("failed to create local storage: %v", err)
	}

	ctx := context.Background()
	pkgName := "testpkg"
	fileName := "file.txt"
	fileContent := "hello world"

	// WriteFile
	writePath := filepath.Join(pkgName, fileName)
	err = storage.WriteFile(ctx, writePath, strings.NewReader(fileContent))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// ListPackages
	pkgs, err := storage.ListPackages(ctx)
	if err != nil {
		t.Fatalf("ListPackages failed: %v", err)
	}
	if len(pkgs) != 1 || pkgs[0] != pkgName {
		t.Errorf("ListPackages got %v, want [%q]", pkgs, pkgName)
	}

	// ListPackageFiles
	files, err := storage.ListPackageFiles(ctx, pkgName)
	if err != nil {
		t.Fatalf("ListPackageFiles failed: %v", err)
	}
	if len(files) != 1 || files[0] != fileName {
		t.Errorf("ListPackageFiles got %v, want [%q]", files, fileName)
	}

	// ReadFile
	reader, err := storage.ReadFile(ctx, writePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(data) != fileContent {
		t.Errorf("ReadFile got %q, want %q", string(data), fileContent)
	}

	// DeleteFile
	err = storage.DeleteFile(ctx, writePath)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}
	// Confirm deletion
	_, err = storage.ReadFile(ctx, writePath)
	if !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted, got err: %v", err)
	}
}
