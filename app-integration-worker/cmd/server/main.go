package server

import (
	"github.com/verasthiago/verancial/app-integration-worker/pkg"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/builder"
	shared "github.com/verasthiago/verancial/shared/flags"
)

func Execute() {
	sharedEnvConfigFile := shared.GetFileEnvConfigFromDeployEnv(shared.SHARED_PACKAGE_NAME)
	builder := new(builder.ServerBuilder).InitBuilder(sharedEnvConfigFile)
	server := new(pkg.Server).InitFromBuilder(builder)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
