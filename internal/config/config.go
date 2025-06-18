package config

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"
)

type Config struct {
	RunAddress string
	BaseURL    string
}

var (
	cfg  *Config
	once sync.Once
)

func Init() {
	once.Do(func() {
		flagRunAddr := flag.String("a", "", "http server run address")
		flagBaseURL := flag.String("b", "", "base url for short address")
		flag.Parse()

		defaultRunAddr := "localhost:8080"
		defaultBaseURL := "http://localhost:8080/"

		runAddr := defaultRunAddr
		baseURL := defaultBaseURL

		if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
			runAddr = envRunAddr
		} else if *flagRunAddr != "" {
			runAddr = *flagRunAddr
		}

		if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
			baseURL = envBaseURL
		} else if *flagBaseURL != "" {
			baseURL = *flagBaseURL
		}

		cfg = &Config{
			RunAddress: runAddr,
			BaseURL:    strings.TrimRight(baseURL, "/"),
		}
	})
}

func Get() *Config {
	if cfg == nil {
		panic("config not initialized: call config.Init() before config.Get()")
	}
	log.Println("Config initiated")
	return cfg
}
