package main

import (
	"context"
	"embed"
	"github.com/exelban/cheks/api"
	"github.com/exelban/cheks/runner"
	"github.com/exelban/cheks/types"
	"github.com/pkg/errors"
	"github.com/pkgz/rest"
	"github.com/pkgz/service"
	"html/template"
	"log"
	"os"
)

type args struct {
	Live       bool   `long:"live" env:"LIVE" description:"live preview of index.html"`
	ConfigPath string `long:"config" env:"CONFIG" default:"./config.yaml" description:"path to the configuration file"`

	DashboardAuth     bool   `long:"dashboard-auth" env:"DASHBOARD_AUTH" description:"secure dashboard with credentials"`
	DashboardUsername string `long:"dashboard-username" env:"DASHBOARD_USERNAME" default:"admin" description:"dashboard username"`
	DashboardPassword string `long:"dashboard-password" env:"DASHBOARD_PASSWORD" description:"dashboard password"`

	service.ARGS
}

type app struct {
	args args

	monitor *runner.Monitor
	api     *api.Rest

	srv *rest.Server
}

//go:embed index.html
var indexHTML embed.FS

const version = "v0.0.0"

func main() {
	var args args
	ctx, _, err := service.Init(&args)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		os.Exit(1)
	}

	app, err := New(args)
	if err != nil {
		log.Printf("[ERROR] setup application failed: %+v", err)
		os.Exit(2)
	}

	log.Print("[INFO] starting an application")

	if err := app.run(ctx); err != nil {
		log.Printf("[ERROR] start application failed: %+v", err)
		os.Exit(3)
	}

	log.Print("[INFO] application terminated")
}

func New(args args) (*app, error) {
	cfg := &types.Config{}
	if err := cfg.Parse(args.ConfigPath); err != nil {
		return nil, errors.Wrap(err, "parse config")
	}

	log.Printf("[INFO] default settings: MaxConn=%d, Retry=%s, Timeout=%s, InitialDelay=%s, Success=%+v, SuccessThreshold=%d, FailureThreshold=%d",
		cfg.MaxConn, cfg.Retry, cfg.Timeout, cfg.InitialDelay, cfg.Success, cfg.SuccessThreshold, cfg.FailureThreshold)

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "validate config")
	}

	monitor := &runner.Monitor{
		Config: cfg,
		Dialer: runner.NewDialer(cfg.MaxConn),
	}

	indexHTMLTemplate, err := template.ParseFS(indexHTML, "index.html")
	if err != nil {
		return nil, errors.Wrap(err, "parse index.html")
	}

	return &app{
		args: args,

		monitor: monitor,
		api: &api.Rest{
			Monitor:  monitor,
			Version:  version,
			Live:     args.Live,
			Template: indexHTMLTemplate,
		},

		srv: &rest.Server{
			Port: args.Port,
		},
	}, nil
}

func (a *app) run(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		if err := a.srv.Shutdown(); err != nil {
			log.Printf("[ERROR] rest shutdown %v", err)
		}
	}()

	if err := a.monitor.Run(ctx); err != nil {
		return errors.Wrap(err, "run monitor")
	}

	return a.srv.Run(a.api.Router())
}
