package cmd

import (
	"github.com/verasthiago/verancial/api/pkg"
	"github.com/verasthiago/verancial/api/pkg/builder"
	shared "github.com/verasthiago/verancial/shared/flags"
)

const API_SERVICE_NAME string = "api"

func Execute() {
	sharedEnvConfigFile := shared.GetFileEnvConfigFromDeployEnv(shared.SHARED_PACKAGE_NAME)
	apiEnvConfigFile := shared.GetFileEnvConfigFromDeployEnv(API_SERVICE_NAME)
	builder := new(builder.ServerBuilder).InitBuilder(apiEnvConfigFile, sharedEnvConfigFile)
	server := new(pkg.Server).InitFromBuilder(builder)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
