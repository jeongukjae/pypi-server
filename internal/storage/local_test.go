package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jeongukjae/pypi-server/internal/config"
)

func TestLocalStorage(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(&config.LocalConfig{Path: dir})

	ctx := context.Background()
	pkgName := "testpkg"
	fileName := "file.txt"
	fileContent := "hello world"

	// WriteFile
	writePath := filepath.Join(pkgName, fileName)
	err := storage.WriteFile(ctx, writePath, strings.NewReader(fileContent))
	require.NoError(t, err)

	// ListPackages
	pkgs, err := storage.ListPackages(ctx)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Equal(t, pkgName, pkgs[0])

	// ListPackageFiles
	files, err := storage.ListPackageFiles(ctx, pkgName)
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, fileName, files[0])

	// ReadFile
	reader, err := storage.ReadFile(ctx, writePath)
	require.NoError(t, err)

	defer reader.Close()
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, string(data), fileContent)

	// DeleteFile
	err = storage.DeleteFile(ctx, writePath)
	require.NoError(t, err)

	// Confirm deletion
	_, err = storage.ReadFile(ctx, writePath)
	assert.True(t, os.IsNotExist(err))
}
