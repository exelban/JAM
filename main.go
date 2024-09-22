package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/exelban/JAM/api"
	"github.com/exelban/JAM/pkg/html"
	"github.com/exelban/JAM/pkg/monitor"
	"github.com/exelban/JAM/store"
	"github.com/exelban/JAM/types"
	"github.com/pkgz/service"
	"log"
	"os"
)

type args struct {
	ConfigPath string `long:"config-path" env:"CONFIG_PATH" default:"./config.yaml" description:"path to the configuration file"`
	service.ARGS
}

type app struct {
	srv *api.Server
	api *api.Rest

	config *types.Cfg
	store  store.Interface

	args args
}

//go:embed templates/*
var fs embed.FS

const version = "v0.0.0"

func main() {
	log.Printf("uptime %s", version)

	var args args
	ctx, _, err := service.Init(&args)
	if err != nil {
		fmt.Printf("error init: %v", err)
		os.Exit(1)
	}

	app, err := create(ctx, args)
	if err != nil {
		log.Printf("[ERROR] create app: %v", err)
		os.Exit(1)
	}

	if err := app.run(ctx); err != nil {
		log.Printf("[ERROR] run app: %v", err)
		os.Exit(1)
	}
}

func create(ctx context.Context, args args) (*app, error) {
	log.Printf("[DEBUG] %+v", args)

	cfg, err := types.NewConfig(ctx, args.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("new config: %w", err)
	}

	storage, err := store.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("new store: %w", err)
	}

	return &app{
		srv: &api.Server{
			Port: args.Port,
		},

		api: &api.Rest{
			Monitor: &monitor.Monitor{
				Store: storage,
			},
			Templates: &html.Templates{
				FS:    fs,
				Debug: args.Debug,
			},
			Version: version,
		},
		config: cfg,
		store:  storage,

		args: args,
	}, nil
}

func (a *app) run(ctx context.Context) error {
	if err := a.api.Templates.Run(ctx); err != nil {
		log.Printf("[ERROR] generate templates: %v", err)
	}

	go func() {
		if err := a.srv.Run(a.api.Router()); err != nil {
			log.Printf("[ERROR] run rest server: %v", err)
		}
	}()

	for {
		select {
		case <-a.config.FW:
			if err := a.config.Parse(); err != nil {
				log.Printf("[ERROR] parse config: %v", err)
			}
			if err := a.config.Validate(); err != nil {
				log.Printf("[ERROR] validate config: %v", err)
			}
			if err := a.api.Monitor.Run(a.config); err != nil {
				log.Printf("[ERROR] run monitor: %v", err)
			}
		case <-ctx.Done():
			log.Print("[DEBUG] terminating...")

			if err := a.srv.Shutdown(); err != nil {
				log.Printf("[ERROR] rest shutdown %v", err)
			}
			if err := a.store.Close(); err != nil {
				log.Printf("[ERROR] store close %v", err)
			}

			return nil
		}
	}
}
