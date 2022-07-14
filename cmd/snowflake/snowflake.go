package snowflake

import (
	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/urfave/cli/v2"
)

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:  "snowflake",
		Usage: "snowflake mod run server",
		Action: func(c *cli.Context) error {
			RegisterService()
			appName := config.GetString("application.name", "com.github.imkuqin_zw.uuid-generator.snowflake")
			return yggdrasil.Run(appName)
		},
	}
}
