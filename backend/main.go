package main

import (
	"context"
	"fmt"
	"github.com/exelban/uptime/api"
	"github.com/exelban/uptime/pkg/monitor"
	"github.com/exelban/uptime/types"
	"github.com/pkgz/rest"
	"github.com/pkgz/service"
	"log"
	"os"
)

type args struct {
	ConfigPath string `long:"config" env:"CONFIG" default:"./config.yaml" description:"path to the configuration file"`

	service.ARGS
}

type app struct {
	srv *rest.Server

	api     *api.Rest
	config  *types.Cfg
	monitor *monitor.Monitor

	args args
}

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
		return nil, err
	}

	mtr := &monitor.Monitor{}

	return &app{
		srv: rest.NewServer(args.Port),

		api: &api.Rest{
			Monitor: mtr,
		},
		config:  cfg,
		monitor: mtr,

		args: args,
	}, nil
}

func (a *app) run(ctx context.Context) error {
	go func() {
		_ = a.srv.Run(a.api.Router(ctx))
	}()

	for {
		select {
		case <-a.config.FW:
			if err := a.config.Parse(); err != nil {
				return fmt.Errorf("parse config: %w", err)
			}
			if err := a.config.Validate(); err != nil {
				return fmt.Errorf("validate config: %w", err)
			}
			if err := a.monitor.Run(a.config); err != nil {
				return fmt.Errorf("reload watcher on config updates: %w", err)
			}
		case <-ctx.Done():
			log.Print("[DEBUG] terminating...")

			if err := a.srv.Shutdown(); err != nil {
				log.Printf("[ERROR] rest shutdown %v", err)
			}
			return nil
		}
	}
}
