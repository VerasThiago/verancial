//go:build mage
// +build mage

package main

import (
	"fmt"
	"time"

	"github.com/verasthiago/verancial/shared/auth"
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/models"
)

func GenerateAdminToken() {
	sharedEnvConfigFile := &shared.EnvFileConfig{
		Path: "../.env",
		Name: fmt.Sprintf("shared.%+v", shared.GetDeployEnv()),
		Type: "env",
	}

	sharedFlags, err := new(shared.SharedFlags).InitFromViper(sharedEnvConfigFile)
	if err != nil {
		panic(err)
	}

	token, err := auth.GenerateJWT(&models.User{IsAdmin: true, IsVerified: true}, sharedFlags.JwtKey, time.Now().Add(10*time.Hour))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Token: %+v\n", token)
}
