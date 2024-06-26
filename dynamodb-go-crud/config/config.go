package config

import (
	"strconv"

	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/utils/env"
)

type Config struct {
	Port int
	Timeout int
	Dialect string
	DatabaseURI string
}

func parseEnvToInt(envName, defaultValue string) int {
	num, err := strconv.Atoi(env.GetEnv(envName, defaultValue))
	if err != nil {
		return 0
	}
	return num
}

func GetConfig() Config {
	return Config{
		Port:		 parseEnvToInt("PORT", "8080"),
		Timeout:	 parseEnvToInt("TIMEOUT", "30"),
		Dialect:	 env.GetEnv("DIALECT", "sqlite3"),
		DatabaseURI: env.GetEnv("DATABASE_URI", ":memory:"),
	}
}
