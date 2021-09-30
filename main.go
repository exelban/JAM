package main

import (
	"context"
	"crypto/rand"
	"embed"
	"github.com/exelban/cheks/api"
	"github.com/exelban/cheks/pkg/monitor"
	"github.com/exelban/cheks/types"
	"github.com/pkg/errors"
	"github.com/pkgz/rest"
	"github.com/pkgz/service"
	"html/template"
	"log"
	"os"
)

type args struct {
	ConfigPath string `long:"config" env:"CONFIG" default:"./config.yaml" description:"path to the configuration file"`

	Auth     bool   `long:"auth" env:"AUTH" description:"secure rest with credentials"`
	Username string `long:"username" env:"USERNAME" default:"admin" description:"username"`
	Password string `long:"password" env:"PASSWORD" description:"password"`

	service.ARGS
}

type app struct {
	args args

	config  *types.Cfg
	monitor *monitor.Monitor
	api     *api.Rest

	srv *rest.Server
}

//go:embed index.html
var indexHTML embed.FS

const version = "v0.0.0"

func main() {
	log.Printf("cheks %s", version)

	var args args
	ctx, _, err := service.Init(&args)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		os.Exit(1)
	}

	app, err := New(ctx, args)
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

func New(ctx context.Context, args args) (*app, error) {
	if args.Auth {
		if args.Username == "" {
			return nil, errors.New("username cannot be empty when AUTH is true")
		}
		if args.Password == "" {
			args.Password = secureRandomAlphaString(32)
			log.Printf("[INFO] password: %s", args.Password)
		}
	}

	cfg, err := types.NewConfig(ctx, args.ConfigPath)
	if err != nil {
		return nil, err
	}

	indexHTMLTemplate, err := template.ParseFS(indexHTML, "index.html")
	if err != nil {
		return nil, errors.Wrap(err, "parse index.html")
	}

	monitor_ := &monitor.Monitor{}

	return &app{
		args: args,

		config:  cfg,
		monitor: monitor_,
		api: &api.Rest{
			Monitor:  monitor_,
			Version:  version,
			Live:     false,
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
		_ = a.srv.Run(a.api.Router())
	}()

	for {
		select {
		case <-a.config.FW:
			if err := a.config.Parse(); err != nil {
				return errors.Wrap(err, "parse config")
			}
			if err := a.config.Validate(); err != nil {
				return errors.Wrap(err, "validate config")
			}
			if err := a.monitor.Run(a.config); err != nil {
				return errors.Wrap(err, "reload watcher on config updates")
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
