package builder

import (
	"testing"

	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestServerBuilder_Getters(t *testing.T) {
	flags := &Flags{Port: "8080"}
	sharedFlags := &shared.SharedFlags{JwtKey: "secret"}
	log := zap.NewNop()

	s := &ServerBuilder{
		Flags:       flags,
		SharedFlags: sharedFlags,
		Log:         log,
	}

	assert.Same(t, flags, s.GetFlags())
	assert.Same(t, sharedFlags, s.GetSharedFlags())
	assert.Same(t, log, s.GetLog())
	assert.Nil(t, s.GetRepository())
	assert.Nil(t, s.GetTask())
	assert.Nil(t, s.GetHTTPClient())
}
