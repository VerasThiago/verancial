package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shared "github.com/verasthiago/verancial/shared/flags"
)

func writeEnvFile(t *testing.T, dir, name string, contents string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".env"), []byte(contents), 0o600))
}

func TestFlags_InitFromViper(t *testing.T) {
	t.Run("loads fields from the env file", func(t *testing.T) {
		dir := t.TempDir()
		writeEnvFile(t, dir, "login.local", "LOGIN_PORT=8081\n")

		flags, err := new(Flags).InitFromViper(&shared.EnvFileConfig{
			Path: dir,
			Name: "login.local",
			Type: "env",
		})

		require.NoError(t, err)
		assert.Equal(t, "8081", flags.Port)
	})

	t.Run("missing config file returns an error", func(t *testing.T) {
		dir := t.TempDir()

		flags, err := new(Flags).InitFromViper(&shared.EnvFileConfig{
			Path: dir,
			Name: "does-not-exist",
			Type: "env",
		})

		assert.Error(t, err)
		assert.Nil(t, flags)
	})

	t.Run("environment variable overrides take effect via AutomaticEnv", func(t *testing.T) {
		dir := t.TempDir()
		writeEnvFile(t, dir, "login.docker", "LOGIN_PORT=9000\n")

		t.Setenv("LOGIN_PORT", "9999")

		flags, err := new(Flags).InitFromViper(&shared.EnvFileConfig{
			Path: dir,
			Name: "login.docker",
			Type: "env",
		})

		require.NoError(t, err)
		assert.Equal(t, "9999", flags.Port)
	})
}
