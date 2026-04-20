package env

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	//Check a known variable to see if env vars are already loaded
	//if not, load godotenv
	if val := GetString("APP_VERSION", "NOT_LOADED"); val == "NOT_LOADED" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Function to read environment variables
func GetString(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Printf("Using fallback value (%s) for key (%s)\n", fallback, key)
		return fallback
	}
	return val
}

func GetInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		log.Printf("Using fallback value (%d) for key (%s)\n", fallback, key)
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Using fallback value (%d) for key (%s)\n", fallback, key)
		return fallback
	}
	return valAsInt
}

func GetDuration(key string, fallback string) time.Duration {
	val := os.Getenv(key)
	var fbpd time.Duration
	var kpd time.Duration
	var err error

	// Try parsing the fallback
	if fbpd, err = time.ParseDuration(fallback); err != nil {
		fbpd = 0
	}

	// Try parsing key
	if kpd, err = time.ParseDuration(val); err != nil {
		kpd = 0
	}

	// If the key was empty use fall back
	if val == "" {
		// If fallback also failed to parse, 0 returned either way
		return fbpd
	}

	return kpd
}
