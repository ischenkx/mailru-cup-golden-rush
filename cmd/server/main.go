package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/RomanIschenko/golden-rush-mailru/internal/app"
	"github.com/RomanIschenko/golden-rush-mailru/internal/config"
	"log"
	_ "net/http/pprof"
	"os"
)

var configPath = flag.String("cfg", "config.json", "Application config")

func main() {
	//runtime.SetBlockProfileRate(1)
	//go http.ListenAndServe("localhost:3434", nil)

	flag.Parse()
	var cfg config.Config

	configFile, err := os.OpenFile(*configPath, os.O_RDONLY, 0755)

	if err != nil {
		log.Println("error while reading config file:", err)
		return
	}

	json.NewDecoder(configFile).Decode(&cfg)

	ADDRESS := os.Getenv("ADDRESS")
	PORT := 8000
	SCHEMA := "http"

	if ADDRESS == "" {
		ADDRESS = "localhost"
	}

	cfg.BaseURL = fmt.Sprintf("http://%s:%v", ADDRESS, PORT)

	log.Printf("STARTING A SERVER (ADDRESS=%s, PORT=%v, SCHEMA=%s)\n", ADDRESS, PORT, SCHEMA)
	app.New(cfg).Start(context.Background())
}