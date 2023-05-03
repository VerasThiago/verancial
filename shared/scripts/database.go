//go:build mage
// +build mage

package main

import (
	"fmt"

	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/models"
	postgresrepository "github.com/verasthiago/verancial/shared/repository/postgresRepository"
)

func MigrateUserModel() {
	sharedEnvConfigFile := &shared.EnvFileConfig{
		Path: "../.env",
		Name: fmt.Sprintf("shared.%+v", shared.GetDeployEnv()),
		Type: "env",
	}

	sharedFlags, err := new(shared.SharedFlags).InitFromViper(sharedEnvConfigFile)
	if err != nil {
		panic(err)
	}

	db := new(postgresrepository.PostgresRepository).InitFromFlags(sharedFlags)

	if err := db.MigrateUser(&models.User{}); err != nil {
		panic(err)
	}

	fmt.Println("User migration done!")
}

func MigrateTransactionModel() {
	sharedEnvConfigFile := &shared.EnvFileConfig{
		Path: "../.env",
		Name: fmt.Sprintf("shared.%+v", shared.GetDeployEnv()),
		Type: "env",
	}

	sharedFlags, err := new(shared.SharedFlags).InitFromViper(sharedEnvConfigFile)
	if err != nil {
		panic(err)
	}

	db := new(postgresrepository.PostgresRepository).InitFromFlags(sharedFlags)

	if err := db.MigrateTransaction(&models.Transaction{}); err != nil {
		panic(err)
	}

	fmt.Println("Transaction migration done!")
}
