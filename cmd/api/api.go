package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/strax84mb/go-travel-reactive/internal/app"
	"github.com/strax84mb/go-travel-reactive/internal/services/auth"
	"github.com/strax84mb/go-travel-reactive/internal/storage"
	"github.com/strax84mb/go-travel-reactive/internal/web/handlers"
	"gopkg.in/yaml.v3"
)

func main() {
	args := parseArgs(os.Args)

	cfg, err := loadConfig(args["config"])
	if err != nil {
		log.Fatalf("could not load config: %s", err.Error())
	}

	ctx := context.Background()
	logger := app.NewLogger()

	repository, err := storage.NewRepository(cfg.API.DbDsn)
	if err != nil {
		logger.Error(app.ContextWithError(ctx, err), "can't connect to DB")
	}

	authentication := auth.NewAuthService(repository, logger)

	r := mux.NewRouter()
	s := r.PathPrefix("/gotravel/reactivex/v1").Subrouter()

	handlers.RegisterTestHandler(s)
	handlers.RegisterUserHandlers(s, authentication)

	if err := http.ListenAndServe(cfg.API.Listen, r); err != nil {
		logger.Error(app.ContextWithError(ctx, err), "can't start server")
	}
}

func parseArgs(args []string) map[string]string {
	if len(args) > 1 {
		args = args[1:]
	}

	m := make(map[string]string)

	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--config="):
			m["config"] = strings.TrimPrefix(a, "--config=")
		}
	}

	if _, ok := m["config"]; !ok {
		m["config"] = "config.yml"
	}

	return m
}

func loadConfig(file string) (*app.Config, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read from file: %w", err)
	}

	cfg := app.Config{}

	if err = yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("could not unmarshall data: %w", err)
	}

	return &cfg, nil
}
