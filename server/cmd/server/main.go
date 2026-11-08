package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anishgowda21/code-forge/server/internal/api"
)

func getStartupEnv() (port, dataDirPath string) {
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataDirPath = os.Getenv("DATA_DIR")
	if dataDirPath == "" {
		dataDirPath = "./data"
	}

	return port, dataDirPath
}

func createDataDir(path string) error {
	err := os.MkdirAll(filepath.Join(path, "repo"), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(path, "db"), 0755)
	if err != nil {
		return err
	}
	return nil
}

func preStartupCheck() (port, dataDirPath string) {
	port, dataDirPath = getStartupEnv()
	log.Printf("Port configured to: %s", port)
	log.Printf("Data folder configured to: %s", dataDirPath)
	err := createDataDir(dataDirPath)
	if err != nil {
		panic(err)
	}
	return port, dataDirPath
}

func main() {
	port, _ := preStartupCheck()

	mux := api.NewServer()

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
