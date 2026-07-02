package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	shared "github.com/verasthiago/verancial/shared/flags"
	repositorymocks "github.com/verasthiago/verancial/shared/repository/mocks"
	"go.uber.org/mock/gomock"
)

// TestServerBuilder_Getters exercises the trivial accessor methods on
// ServerBuilder. InitBuilder is intentionally not covered here: it reads
// real env config files via viper and opens a live Postgres connection
// (or panics), so it isn't meaningfully unit-testable without turning it
// into an integration test against a real/dockerized Postgres - out of
// scope for this package's unit tests.
func TestServerBuilder_Getters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flags := &Flags{Port: "8080"}
	sharedFlags := &shared.SharedFlags{JwtKey: "key"}
	repo := repositorymocks.NewMockRepository(ctrl)
	log := zap.NewNop()

	sb := &ServerBuilder{
		Flags:       flags,
		SharedFlags: sharedFlags,
		Repository:  repo,
		Log:         log,
	}

	assert.Same(t, flags, sb.GetFlags())
	assert.Same(t, sharedFlags, sb.GetSharedFlags())
	assert.Same(t, log, sb.GetLog())
	assert.Equal(t, repo, sb.GetRepository())
}
