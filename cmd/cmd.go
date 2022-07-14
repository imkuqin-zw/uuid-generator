package cmd

import (
	"github.com/imkuqin-zw/uuid-generator/cmd/segment"
	"github.com/imkuqin-zw/uuid-generator/cmd/snowflake"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	"github.com/urfave/cli/v2"
)

func NewServerCmd() *cli.App {
	app := &cli.App{
		Name:  "uuid-generator",
		Usage: "A global UUID generation application",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "config file path",
				Value:   "./config/config.yaml",
				EnvVars: []string{"UUID_GENERATOR_CONFIG"},
			},
		},
		Before: func(c *cli.Context) error {
			err := config.LoadSource(file.NewSource(c.String("config"), false))
			if err != nil {
				return err
			}
			return nil
		},
	}
	registerCommands(app)
	return app
}

func registerCommands(app *cli.App) {
	svrCmd := &cli.Command{
		Name:  "serve",
		Usage: "run server",
		Subcommands: []*cli.Command{
			segment.NewCmd(),
			snowflake.NewCmd(),
		},
	}
	app.Commands = append(app.Commands, svrCmd)
}
