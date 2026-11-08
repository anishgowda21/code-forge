package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anishgowda21/code-forge/server/internal/api"
	"github.com/anishgowda21/code-forge/server/internal/database"
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

func createDataDir(path string) (dbPath, repoPath string, err error) {
	dbPath = filepath.Join(path, "db")
	err = os.MkdirAll(dbPath, 0755)
	if err != nil {
		return "", "", err
	}

	repoPath = filepath.Join(path, "repo")
	err = os.MkdirAll(dbPath, 0755)
	if err != nil {
		return "", "", err
	}

	return dbPath, repoPath, nil
}

func preStartupCheck() (port, dbPath, repoPath string) {
	port, dataDirPath := getStartupEnv()
	log.Printf("Port configured to: %s", port)
	log.Printf("Data folder configured to: %s", dataDirPath)
	dbPath, repoPath, err := createDataDir(dataDirPath)
	if err != nil {
		log.Fatal(err)
	}
	return port, dbPath, repoPath
}

func main() {
	port, dbPath, _ := preStartupCheck()
	db, err := database.OpenDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := api.NewServer(db)

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
