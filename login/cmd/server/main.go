package cmd

import (
	"github.com/verasthiago/verancial/login/pkg"
	"github.com/verasthiago/verancial/login/pkg/builder"
	shared "github.com/verasthiago/verancial/shared/flags"
)

const LOGIN_SERVICE_NAME string = "login"

func Execute() {
	sharedEnvConfigFile := shared.GetFileEnvConfigFromDeployEnv(shared.SHARED_PACKAGE_NAME)
	loginEnvConfigFile := shared.GetFileEnvConfigFromDeployEnv(LOGIN_SERVICE_NAME)
	builder := new(builder.ServerBuilder).InitBuilder(loginEnvConfigFile, sharedEnvConfigFile)
	server := new(pkg.Server).InitFromBuilder(builder)

	if err := server.Run(); err != nil {
		panic(err)
	}
}
