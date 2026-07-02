package builder

import (
	"os"
	"path/filepath"
	"testing"

	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerBuilder_InitBuilder_PanicsOnUnreachableDB exercises InitBuilder's
// real wiring (Flags.InitFromViper, SharedFlags.InitFromViper, zap logger)
// through valid temp config files, up to the point where it opens a
// Postgres connection. PostgresRepository.InitFromFlags panics on
// connection failure rather than returning an error, so a bad DSN pointing
// at an unreachable host is the only way to observe this path complete
// without a live database -- InitBuilder is fail-fast by design.
func TestServerBuilder_InitBuilder_PanicsOnUnreachableDB(t *testing.T) {
	loginDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(loginDir, "login.test.env"), []byte("LOGIN_PORT=8081\n"), 0o600))

	sharedDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sharedDir, "shared.test.env"), []byte(
		"DB_HOST=127.0.0.1\nDB_PORT=1\nDB_USER=postgres\nDB_PASSWORD=postgres\nDB_NAME=postgres\nDB_SSLMODE=disable\nDB_TIMEZONE=UTC\nJWT_KEY=test\n",
	), 0o600))

	loginEnvConfig := &shared.EnvFileConfig{Path: loginDir, Name: "login.test", Type: "env"}
	sharedEnvConfig := &shared.EnvFileConfig{Path: sharedDir, Name: "shared.test", Type: "env"}

	assert.Panics(t, func() {
		new(ServerBuilder).InitBuilder(loginEnvConfig, sharedEnvConfig)
	}, "InitBuilder should panic when it can't reach Postgres, not swallow the error")
}

func TestServerBuilder_InitBuilder_PanicsOnMissingConfigFile(t *testing.T) {
	emptyDir := t.TempDir()

	missingConfig := &shared.EnvFileConfig{Path: emptyDir, Name: "does-not-exist", Type: "env"}

	assert.Panics(t, func() {
		new(ServerBuilder).InitBuilder(missingConfig, missingConfig)
	}, "InitBuilder should panic when the login env config file can't be read, not swallow the error")
}
