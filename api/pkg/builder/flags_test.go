package builder

import (
	"os"
	"path/filepath"
	"testing"

	shared "github.com/verasthiago/verancial/shared/flags"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeEnvFixture(t *testing.T, dir, fileName, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0600)
	require.NoError(t, err)
}

func TestFlags_InitFromViper(t *testing.T) {
	t.Run("success reads API_PORT and WEB_PORT from a fixture file", func(t *testing.T) {
		dir := t.TempDir()
		writeEnvFixture(t, dir, "api.test.env", "API_PORT=8080\nWEB_PORT=3000\n")

		cfg := &shared.EnvFileConfig{
			Path: dir,
			Name: "api.test",
			Type: "env",
		}

		flags, err := new(Flags).InitFromViper(cfg)

		require.NoError(t, err)
		require.NotNil(t, flags)
		assert.Equal(t, "8080", flags.Port)
		assert.Equal(t, "3000", flags.WebPort)
	})

	t.Run("missing config file returns an error", func(t *testing.T) {
		dir := t.TempDir()

		cfg := &shared.EnvFileConfig{
			Path: dir,
			Name: "does-not-exist",
			Type: "env",
		}

		flags, err := new(Flags).InitFromViper(cfg)

		require.Error(t, err)
		assert.Nil(t, flags)
	})

	t.Run("malformed config file returns an error", func(t *testing.T) {
		dir := t.TempDir()
		// Invalid syntax for an env file (no key=value, unterminated quote).
		writeEnvFixture(t, dir, "bad.test.env", `API_PORT="unterminated`)

		cfg := &shared.EnvFileConfig{
			Path: dir,
			Name: "bad.test",
			Type: "env",
		}

		_, err := new(Flags).InitFromViper(cfg)

		require.Error(t, err)
	})

	t.Run("partial fixture leaves missing fields as zero value", func(t *testing.T) {
		dir := t.TempDir()
		writeEnvFixture(t, dir, "partial.test.env", "API_PORT=9090\n")

		cfg := &shared.EnvFileConfig{
			Path: dir,
			Name: "partial.test",
			Type: "env",
		}

		flags, err := new(Flags).InitFromViper(cfg)

		require.NoError(t, err)
		assert.Equal(t, "9090", flags.Port)
		assert.Equal(t, "", flags.WebPort)
	})
}
