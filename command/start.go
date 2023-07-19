package command

import (
	"context"
	"fmt"
	"github.com/puras/mog/config"
	"github.com/puras/mog/logger"
	"github.com/puras/mog/server"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
)

func StartCmd(params server.ServerParam) *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "Start Server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "conf",
				Aliases: []string{"c"},
				Usage:   "Configuration file",
				Value:   "conf/config.toml",
			},
			&cli.BoolFlag{
				Name:    "daemon",
				Aliases: []string{"d"},
				Usage:   "Run as a daemon",
			},
		},
		Action: func(c *cli.Context) error {
			defer func() {
				_ = zap.L().Sync()
			}()
			ctx := logger.NewTag(context.Background(), logger.TagKeyMain)
			return server.Run(ctx, func(ctx context.Context) (func(), error) {
				confFile := c.String("conf")
				daemon := c.Bool("daemon")

				if daemon {
					bin, err := filepath.Abs(os.Args[0])
					if err != nil {
						logger.Context(ctx).Error("Failed to get absolute path for command", zap.Error(err))
						return nil, err
					}
					command := exec.Command(bin, "start", "--conf", confFile)
					err = command.Start()
					if err != nil {
						logger.Context(ctx).Error("Failed to start daemon thread", zap.Error(err))
						return nil, err
					}
					_ = os.WriteFile(fmt.Sprintf("%s.lock", c.App.Name), []byte(fmt.Sprintf("%d", command.Process.Pid)), 0666)
					os.Exit(0)
				}

				config.MustLoad(confFile)
				if config.C.IsDebug() {
					config.C.Print()
				}
				loggerClean, err := logger.Init(ctx)
				if err != nil {
					return nil, err
				}
				logger.Context(ctx).Info("Starting server",
					zap.String("config file", confFile),
					zap.Bool("daemon", daemon),
					zap.Int("pid", os.Getpid()),
				)

				if err := params.Init(ctx); err != nil {
					return nil, err
				}

				startClean, err := server.Start(ctx, params.GetInjector(ctx), params.RegistryRoutes, params.ParseCurrentUser)
				if err != nil {
					return nil, err
				}

				return func() {
					if startClean != nil {
						startClean()
					}
					if loggerClean != nil {
						loggerClean()
					}
				}, nil
			})
		},
	}
}
