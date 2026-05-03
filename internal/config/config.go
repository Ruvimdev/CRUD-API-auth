package config

import "os"


type Config struct {
	DBHost     string 
	DBUser     string
	DBName     string
	DBPassword string
	DBPort     string 
	AppPort    string
}

func LoadConfig() *Config {
	return &Config{
		DBHost: os.Getenv("DB_HOST"),
		DBUser: os.Getenv("DB_USER"),
		DBName: os.Getenv("DB_NAME"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBPort: os.Getenv("DB_PORT"),
		AppPort: os.Getenv("APP_PORT"),

	}
}