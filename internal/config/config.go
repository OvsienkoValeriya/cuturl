package config

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"
)

type Config struct {
	RunAddress      string
	BaseURL         string
	FileStoragePath string
}

var (
	cfg  *Config
	once sync.Once
)

func Init() {
	once.Do(func() {
		flagRunAddr := flag.String("a", "", "http server run address")
		flagBaseURL := flag.String("b", "", "base url for short address")
		flagFileStoragePath := flag.String("f", "", "path for file storage")
		flag.Parse()

		defaultRunAddr := "localhost:8080"
		defaultBaseURL := "http://localhost:8080/"
		defaultFileStoragePath := "/tmp/urls.json"

		runAddr := defaultRunAddr
		baseURL := defaultBaseURL
		fileStoragePath := defaultFileStoragePath

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

		if envFileStorage := os.Getenv("FILE_STORAGE_PATH"); envFileStorage != "" {
			fileStoragePath = envFileStorage
		} else if *flagFileStoragePath != "" {
			fileStoragePath = *flagFileStoragePath
		}

		cfg = &Config{
			RunAddress:      runAddr,
			BaseURL:         strings.TrimRight(baseURL, "/"),
			FileStoragePath: strings.TrimRight(fileStoragePath, "/"),
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
