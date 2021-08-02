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
	"net/http"
	"os"
)

type args struct {
	Live       bool   `long:"live" env:"LIVE" description:"live preview of index.html"`
	ConfigPath string `long:"config" env:"CONFIG" default:"./config.yaml" description:"path to the configuration file"`

	MaxConn int    `long:"max-conn" env:"MAX_CONN" default:"32" description:"maximum parallel request"`
	Retry   string `long:"retry" env:"RETRY" default:"30s" description:"default retry interval"`
	Timeout string `long:"timeout" env:"TIMEOUT" default:"180s" description:"default request timeout"`

	InitialDelay     string `long:"initial-delay" env:"INITIAL-DELAY" default:"0" description:"default initial delay"`
	SuccessCode      []int  `long:"success-code" env:"SUCCESS-CODE" description:"default success codes"`
	SuccessThreshold int    `long:"success-threshold" env:"SUCCESS-THRESHOLD" default:"2" description:"default success threshold"`
	FailureThreshold int    `long:"failure-threshold" env:"FAILURE-THRESHOLD" default:"3" description:"default failure threshold"`

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

	if cfg.MaxConn == 0 {
		cfg.MaxConn = args.MaxConn
	}
	if cfg.Retry == "" {
		cfg.Retry = args.Retry
	}
	if cfg.Timeout == "" {
		cfg.Timeout = args.Timeout
	}
	if cfg.InitialDelay == "" {
		cfg.InitialDelay = args.InitialDelay
	}
	if len(cfg.SuccessCode) == 0 {
		cfg.SuccessCode = args.SuccessCode
		if len(cfg.SuccessCode) == 0 {
			cfg.SuccessCode = []int{
				http.StatusOK,
				http.StatusCreated,
				http.StatusAccepted,
				http.StatusNonAuthoritativeInfo,
				http.StatusNoContent,
				http.StatusResetContent,
				http.StatusPartialContent,
				http.StatusMultiStatus,
				http.StatusAlreadyReported,
				http.StatusIMUsed,
			}
		}
	}
	if cfg.SuccessThreshold == 0 {
		cfg.SuccessThreshold = args.SuccessThreshold
	}
	if cfg.FailureThreshold == 0 {
		cfg.FailureThreshold = args.FailureThreshold
	}

	log.Printf("[INFO] default settings: MaxConn=%d, Retry=%s, Timeout=%s, InitialDelay=%s, SuccessCode=%v, SuccessThreshold=%d, FailureThreshold=%d",
		cfg.MaxConn, cfg.Retry, cfg.Timeout, cfg.InitialDelay, cfg.SuccessCode, cfg.SuccessThreshold, cfg.FailureThreshold)

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "validate config")
	}

	monitor := &runner.Monitor{
		Config: cfg,
		Dialer: runner.NewDialer(args.MaxConn),
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
