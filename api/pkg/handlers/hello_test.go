package handlers

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"
	buildermocks "github.com/verasthiago/verancial/api/pkg/builder/mocks"
	"github.com/verasthiago/verancial/api/pkg/testutil"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelloHandler_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)

	c, w := testutil.NewGinContext("GET", "/api/v0/hello", nil, "")

	handler := new(HelloHandler).InitFromBuilder(mockBuilder)
	err := handler.Handler(c)

	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "success", body["status"])
	assert.Equal(t, "hello!", body["message"])
}
