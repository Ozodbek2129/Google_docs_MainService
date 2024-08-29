package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	GOOGLE_DOCS    string
	MongoURI          string
	MongoDBName       string
}

func Load() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Print("No .env file found")
	}

	config := Config{}
	config.GOOGLE_DOCS = cast.ToString(Coalesce("GOOGLE_DOCS", ":50052"))
	config.MongoURI = cast.ToString(Coalesce("MONGO_URI", "mongodb://localhost:27017"))
	config.MongoDBName = cast.ToString(Coalesce("MONGODB_NAME", "google_docs"))

	return config
}

func Coalesce(key string, defaultValue interface{}) interface{} {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
