package config

import "go-saga-pattern/commoner/utils"

const (
	EnvLocal       = "local"
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

var environtment string

func init() {
	environtment = utils.GetEnv("ENVIRONMENT")
}

func Get() string {
	return environtment
}

func IsLocal() bool {
	return Get() == EnvLocal
}

func IsDevelopment() bool {
	return Get() == EnvDevelopment
}

func IsProduction() bool {
	return Get() == EnvProduction
}
