package main

import (
	"context"
	"crypto/rand"
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

	Auth     bool   `long:"auth" env:"AUTH" description:"secure rest with credentials"`
	Username string `long:"username" env:"USERNAME" default:"admin" description:"username"`
	Password string `long:"password" env:"PASSWORD" description:"password"`

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
	if args.Auth {
		if args.Username == "" {
			return nil, errors.New("username cannot be empty when DASHBOARD_AUTH is true")
		}
		if args.Password == "" {
			args.Password = secureRandomAlphaString(32)
			log.Printf("[INFO] automatically generate password: %s", args.Password)
		}
	}

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
			Auth: api.Auth{
				Enabled:  args.Auth,
				Username: args.Username,
				Password: args.Password,
			},
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

func secureRandomAlphaString(length int) string {
	const (
		letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // 52 possibilities
		letterIdxBits = 6                                                                // 6 bits to represent 64 possibilities / indexes
		letterIdxMask = 1<<letterIdxBits - 1                                             // All 1-bits, as many as letterIdxBits
	)

	result := make([]byte, length)
	bufferSize := int(float64(length) * 1.3)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes = secureRandomBytes(bufferSize)
		}
		if idx := int(randomBytes[j%length] & letterIdxMask); idx < len(letterBytes) {
			result[i] = letterBytes[idx]
			i++
		}
	}

	return string(result)
}
func secureRandomBytes(length int) []byte {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)

	if err != nil {
		log.Fatal("Unable to generate random bytes")
	}
	return randomBytes
}
