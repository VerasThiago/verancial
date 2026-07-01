package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	buildermocks "github.com/verasthiago/verancial/login/pkg/builder/mocks"
	"github.com/verasthiago/verancial/login/pkg/builder"
	sharedflags "github.com/verasthiago/verancial/shared/flags"
)

func TestServer_InitFromBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockBuilder.EXPECT().GetFlags().Return(&builder.Flags{Port: "8080"}).AnyTimes()
	mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: "key"}).AnyTimes()

	server := new(Server).InitFromBuilder(mockBuilder)

	require.NotNil(t, server)
	assert.NotNil(t, server.LoginAPI)
	assert.NotNil(t, server.CreateAPI)
	assert.NotNil(t, server.DeleteAPI)
	assert.NotNil(t, server.UpdateAPI)
	assert.NotNil(t, server.AdminAPI)
	assert.Equal(t, builder.Builder(mockBuilder), server.Builder)
}
