package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress string
	BaseURL    string
}

var cfg *Config

func Init() {
	flagRunAddr := flag.String("a", "", "http server run address")
	flagBaseUrl := flag.String("b", "", "base url for short address")
	flag.Parse()

	defaultRunAddr := "localhost:8080"
	defaultBaseUrl := "http://localhost:8080/"

	runAddr := defaultRunAddr
	baseUrl := defaultBaseUrl

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		runAddr = envRunAddr
	} else if *flagRunAddr != "" {
		runAddr = *flagRunAddr
	}

	if envBaseUrl := os.Getenv("BASE_URL"); envBaseUrl != "" {
		baseUrl = envBaseUrl
	} else if *flagBaseUrl != "" {
		baseUrl = *flagBaseUrl
	}

	cfg = &Config{
		RunAddress: runAddr,
		BaseURL:    baseUrl,
	}
}

func Get() *Config {
	return cfg
}
