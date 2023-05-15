package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func EnvVariablesInit() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

func Get_Env_Variable(key string) string {

	value, present := os.LookupEnv(key)
	if present == false {
		log.Fatalf("Environment variable %v not present.", key)
	}

	return value
}
