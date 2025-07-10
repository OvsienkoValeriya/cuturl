package config

import (
	"flag"
	"log"
	"os"
	"sync"
)

type Config struct {
	RunAddress      string
	BaseURL         string
	FileStoragePath string
	DBConnection    string
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
		flagDBConnection := flag.String("d", "", "database connection string")
		flag.Parse()

		defaultRunAddr := "localhost:8080"
		defaultBaseURL := "http://localhost:8080/"
		defaultFileStoragePath := "/tmp/urls.json"
		defaultDBConnection := ""

		runAddr := defaultRunAddr
		baseURL := defaultBaseURL
		fileStoragePath := defaultFileStoragePath
		dbConnection := defaultDBConnection

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

		if envDBConnection := os.Getenv("DATABASE_DSN"); envDBConnection != "" {
			dbConnection = envDBConnection
		} else if *flagDBConnection != "" {
			dbConnection = *flagDBConnection
		}

		cfg = &Config{
			RunAddress:      runAddr,
			BaseURL:         baseURL,
			FileStoragePath: fileStoragePath,
			DBConnection:    dbConnection,
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
