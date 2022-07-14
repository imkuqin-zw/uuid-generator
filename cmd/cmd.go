// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
