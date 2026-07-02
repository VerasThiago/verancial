package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	shared "github.com/verasthiago/verancial/shared/flags"
)

func writeConfigFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	fullPath := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	return fullPath
}

func TestFlags_InitFromViper_Success(t *testing.T) {
	dir := t.TempDir()
	writeConfigFile(t, dir, "aiw.test.env", "ASYNC_PROCESSING=true\nAIW_PORT=8080\n")

	config := &shared.EnvFileConfig{
		Path: dir,
		Name: "aiw.test.env",
		Type: "env",
	}

	flags, err := new(Flags).InitFromViper(config)

	require.NoError(t, err)
	require.NotNil(t, flags)
	assert.True(t, flags.AsyncProcessing)
	assert.Equal(t, "8080", flags.Port)
}

func TestFlags_InitFromViper_FalseAsyncProcessing(t *testing.T) {
	dir := t.TempDir()
	writeConfigFile(t, dir, "aiw.test.env", "ASYNC_PROCESSING=false\nAIW_PORT=9090\n")

	config := &shared.EnvFileConfig{
		Path: dir,
		Name: "aiw.test.env",
		Type: "env",
	}

	flags, err := new(Flags).InitFromViper(config)

	require.NoError(t, err)
	assert.False(t, flags.AsyncProcessing)
	assert.Equal(t, "9090", flags.Port)
}

func TestFlags_InitFromViper_FileNotFound(t *testing.T) {
	dir := t.TempDir()

	config := &shared.EnvFileConfig{
		Path: dir,
		Name: "does-not-exist.env",
		Type: "env",
	}

	flags, err := new(Flags).InitFromViper(config)

	assert.Error(t, err)
	assert.Nil(t, flags)
}

func TestFlags_InitFromViper_InvalidDirectory(t *testing.T) {
	config := &shared.EnvFileConfig{
		Path: filepath.Join(t.TempDir(), "does-not-exist-dir"),
		Name: "aiw.test.env",
		Type: "env",
	}

	flags, err := new(Flags).InitFromViper(config)

	assert.Error(t, err)
	assert.Nil(t, flags)
}
