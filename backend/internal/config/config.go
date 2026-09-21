package config

import "os"

// Config holds every runtime knob the API reads from the environment.
type Config struct {
	Addr        string
	AllowOrigin string
}

// Load reads the configuration, falling back to local development defaults.
func Load() Config {
	return Config{
		Addr:        env("DUOTASTE_ADDR", ":8080"),
		AllowOrigin: env("DUOTASTE_ALLOW_ORIGIN", "http://localhost:5173"),
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
