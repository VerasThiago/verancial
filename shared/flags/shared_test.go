package flags

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeEnvFile(t *testing.T, dir, name string, contents string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".env"), []byte(contents), 0o600))
}

func TestSharedFlags_InitFromViper(t *testing.T) {
	t.Run("loads fields from the env file and stamps the deploy env", func(t *testing.T) {
		t.Setenv(DEPLOY_ENV, DEPLOY_LOCAL)

		dir := t.TempDir()
		writeEnvFile(t, dir, "shared.local", "DB_HOST=localhost\nDB_PORT=5432\nJWT_KEY=secret\n")

		flags, err := new(SharedFlags).InitFromViper(&EnvFileConfig{
			Path: dir,
			Name: "shared.local",
			Type: ENV_FILE_TYPE,
		})

		require.NoError(t, err)
		assert.Equal(t, "localhost", flags.DatabaseHost)
		assert.Equal(t, "5432", flags.DatabasePort)
		assert.Equal(t, "secret", flags.JwtKey)
		assert.Equal(t, DEPLOY_LOCAL, flags.Deploy)
	})

	t.Run("missing config file returns an error", func(t *testing.T) {
		dir := t.TempDir()

		_, err := new(SharedFlags).InitFromViper(&EnvFileConfig{
			Path: dir,
			Name: "does-not-exist",
			Type: ENV_FILE_TYPE,
		})

		assert.Error(t, err)
	})
}

func TestGetDeployEnv(t *testing.T) {
	t.Run("valid deploy env is returned", func(t *testing.T) {
		t.Setenv(DEPLOY_ENV, DEPLOY_PRODUCTION)
		assert.Equal(t, DEPLOY_PRODUCTION, GetDeployEnv())
	})

	t.Run("invalid deploy env panics", func(t *testing.T) {
		t.Setenv(DEPLOY_ENV, "not-a-real-env")
		assert.Panics(t, func() { GetDeployEnv() })
	})
}

func TestGetFileEnvConfigFromDeployEnv(t *testing.T) {
	t.Run("shared package uses the shared-relative path", func(t *testing.T) {
		t.Setenv(DEPLOY_ENV, DEPLOY_LOCAL)

		cfg := GetFileEnvConfigFromDeployEnv(SHARED_PACKAGE_NAME)

		assert.Equal(t, "../shared/.env", cfg.Path)
		assert.Equal(t, "shared.local.env", cfg.Name)
		assert.Equal(t, ENV_FILE_TYPE, cfg.Type)
	})

	t.Run("other services use the default .env path", func(t *testing.T) {
		t.Setenv(DEPLOY_ENV, DEPLOY_DOCKER)

		cfg := GetFileEnvConfigFromDeployEnv("api")

		assert.Equal(t, ENV_FILE_PATH, cfg.Path)
		assert.Equal(t, "api.docker.env", cfg.Name)
	})
}
