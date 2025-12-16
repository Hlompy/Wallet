package config

import (
	"os"
)

type Config struct {
	AppPort string
	DBDsn   string
}

func Load() *Config {
	return &Config{
		AppPort: os.Getenv("APP_PORT"),
		DBDsn: "host=" + os.Getenv("DB_HOST") +
			" port=" + os.Getenv("DB_PORT") +
			" user=" + os.Getenv("DB_USER") +
			" password=" + os.Getenv("DB_PASSWORD") +
			" dbname=" + os.Getenv("DB_NAME") +
			" sslmode=" + os.Getenv("DB_SSLMODE"),
	}
}
