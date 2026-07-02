package builder

import (
	"testing"

	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators/mocks"
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestServerBuilder_Getters(t *testing.T) {
	ctrl := gomock.NewController(t)

	flags := &Flags{Port: "8080"}
	sharedFlags := &shared.SharedFlags{JwtKey: "secret"}
	factory := mocks.NewMockAppReportGeneratorFactory(ctrl)

	s := &ServerBuilder{
		Flags:                     flags,
		SharedFlags:               sharedFlags,
		AppReportGeneratorFactory: factory,
	}

	assert.Same(t, flags, s.GetFlags())
	assert.Same(t, sharedFlags, s.GetSharedFlags())
	assert.Nil(t, s.GetRepository())
	assert.Equal(t, factory, s.GetAppReportGeneratorFactory())
}
