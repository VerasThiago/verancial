package server

import (
	"github.com/verasthiago/verancial/data-process-worker/pkg"
	"github.com/verasthiago/verancial/data-process-worker/pkg/builder"
	shared "github.com/verasthiago/verancial/shared/flags"
)

const DATA_PROCESS_SERVICE_NAME string = "dpw"

func Execute() {
	sharedEnvConfigFile := shared.GetFileEnvConfigFromDeployEnv(shared.SHARED_PACKAGE_NAME)
	dataProcessEnvConfigFile := shared.GetFileEnvConfigFromDeployEnv(DATA_PROCESS_SERVICE_NAME)

	builder := new(builder.ServerBuilder).InitBuilder(dataProcessEnvConfigFile, sharedEnvConfigFile)
	server := new(pkg.Server).InitFromBuilder(builder)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
