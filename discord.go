package main

import (
	"os"

	"github.com/joho/godotenv"
)

func grabToken() string {
	err := godotenv.Load()
	HandleErr(err)
	tkn := os.Getenv("token")
	return tkn
}
