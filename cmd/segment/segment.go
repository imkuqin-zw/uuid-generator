package segment

import (
	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/urfave/cli/v2"
)

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:  "segment",
		Usage: "segment mod run server",
		Action: func(c *cli.Context) error {
			RegisterService()
			appName := config.GetString("application.name", "com.github.imkuqin_zw.uuid-generator.segment")
			return yggdrasil.Run(appName)
		},
	}
}
