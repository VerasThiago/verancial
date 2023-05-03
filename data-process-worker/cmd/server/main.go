package server

import (
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/worker/pkg"
	"github.com/verasthiago/verancial/worker/pkg/builder"
)

func Execute() {
	sharedEnvConfigFile := shared.GetFileEnvConfigFromDeployEnv(shared.SHARED_PACKAGE_NAME)
	builder := new(builder.ServerBuilder).InitBuilder(sharedEnvConfigFile)
	server := new(pkg.Server).InitFromBuilder(builder)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
