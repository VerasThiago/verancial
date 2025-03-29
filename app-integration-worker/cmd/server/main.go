package server

import (
	"github.com/verasthiago/verancial/app-integration-worker/pkg"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/builder"
	shared "github.com/verasthiago/verancial/shared/flags"
)

const APP_INTEGRATION_SERVICE_NAME string = "aiw"

func Execute() {
	sharedEnvConfigFile := shared.GetFileEnvConfigFromDeployEnv(shared.SHARED_PACKAGE_NAME)
	appIntegrationConfigFile := shared.GetFileEnvConfigFromDeployEnv(APP_INTEGRATION_SERVICE_NAME)

	builder := new(builder.ServerBuilder).InitBuilder(appIntegrationConfigFile, sharedEnvConfigFile)
	server := new(pkg.Server).InitFromBuilder(builder)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
