package builder

import (
	"testing"

	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/data-process-worker/pkg/report/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestServerBuilder_Getters(t *testing.T) {
	ctrl := gomock.NewController(t)

	flags := &Flags{Port: "8080"}
	sharedFlags := &shared.SharedFlags{JwtKey: "secret"}
	factory := mocks.NewMockReportProcessorFactory(ctrl)

	s := &ServerBuilder{
		Flags:                  flags,
		SharedFlags:            sharedFlags,
		ReportProcessorFactory: factory,
	}

	assert.Same(t, flags, s.GetFlags())
	assert.Same(t, sharedFlags, s.GetSharedFlags())
	assert.Nil(t, s.GetRepository())
	assert.Equal(t, factory, s.GetReportProcessorFactory())
}
