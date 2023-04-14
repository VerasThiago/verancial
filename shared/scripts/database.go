//go:build mage
// +build mage

package main

import (
	"fmt"

	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/models"
	postgresrepository "github.com/verasthiago/verancial/shared/repository/postgresRepository"
)

func MigrateUserDatabase() {
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

	db.MigrateUser(&models.User{})

	fmt.Println("Migration done!")
}

func MigrateApiDatabase() {
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

	// Add API migration models
	fmt.Printf("\ndb %+v\n", db)

	fmt.Println("Migration done!")
}
