package main

import (
	"context"
	"fmt"
	"github.com/exelban/uptime/api"
	"github.com/pkgz/rest"
	"github.com/pkgz/service"
	"log"
	"os"
)

type args struct {
	service.ARGS
}

type app struct {
	srv *rest.Server

	api *api.Rest

	args args
}

const version = "v0.0.0"

func main() {
	log.Printf("cheks %s", version)

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

	return &app{
		srv: rest.NewServer(args.Port),

		api: &api.Rest{},

		args: args,
	}, nil
}

func (a *app) run(ctx context.Context) error {
	return a.srv.Run(a.api.Router(ctx))
}
