package builder

import (
	"os"
	"path/filepath"
	"testing"

	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/stretchr/testify/assert"
)

// TestServerBuilder_InitBuilder_PanicsOnUnreachableDB exercises InitBuilder's
// real wiring (Flags.InitFromViper, SharedFlags.InitFromViper, zap logger,
// AsyncQueue client construction) through valid temp config files, up to the
// point where it opens a Postgres connection. PostgresRepository.InitFromFlags
// panics on connection failure rather than returning an error, so a bad DSN
// pointing at an unreachable host is the only way to observe this path
// complete without a live database -- InitBuilder is fail-fast by design.
func TestServerBuilder_InitBuilder_PanicsOnUnreachableDB(t *testing.T) {
	apiDir := t.TempDir()
	require := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	require(os.WriteFile(filepath.Join(apiDir, "api.test.env"), []byte("API_PORT=8080\nWEB_PORT=3000\n"), 0o600))

	sharedDir := t.TempDir()
	require(os.WriteFile(filepath.Join(sharedDir, "shared.test.env"), []byte(
		"DB_HOST=127.0.0.1\nDB_PORT=1\nDB_USER=postgres\nDB_PASSWORD=postgres\nDB_NAME=postgres\nDB_SSLMODE=disable\nDB_TIMEZONE=UTC\nQUEUE_HOST=127.0.0.1\nQUEUE_PORT=1\nJWT_KEY=test\n",
	), 0o600))

	apiEnvConfig := &shared.EnvFileConfig{Path: apiDir, Name: "api.test", Type: "env"}
	sharedEnvConfig := &shared.EnvFileConfig{Path: sharedDir, Name: "shared.test", Type: "env"}

	assert.Panics(t, func() {
		new(ServerBuilder).InitBuilder(apiEnvConfig, sharedEnvConfig)
	}, "InitBuilder should panic when it can't reach Postgres, not swallow the error")
}

func TestServerBuilder_InitBuilder_PanicsOnMissingConfigFile(t *testing.T) {
	emptyDir := t.TempDir()

	missingConfig := &shared.EnvFileConfig{Path: emptyDir, Name: "does-not-exist", Type: "env"}

	assert.Panics(t, func() {
		new(ServerBuilder).InitBuilder(missingConfig, missingConfig)
	}, "InitBuilder should panic when the api env config file can't be read, not swallow the error")
}
