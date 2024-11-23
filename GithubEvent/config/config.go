package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	BotToken = os.Getenv("TOKEN")
	Port     = os.Getenv("PORT")
)
