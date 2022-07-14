package main

import (
	"os"

	"github.com/imkuqin-zw/uuid-generator/cmd"
	_ "github.com/imkuqin-zw/yggdrasil-gorm/driver/mysql"
	_ "github.com/imkuqin-zw/yggdrasil-gorm/plugin/trace"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/logger/zap"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/polaris/grpc"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/promethues"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/trace/jaeger"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/env"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/server/governor"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/server/grpc/trace"
)

const (
	appName = "com.github.imkuqin_zw.uuid-generator"
)

func main() {
	initConfig()
	app := cmd.NewServerCmd()
	if err := app.Run(os.Args); err != nil {
		os.Exit(-1)
	}
}

func initConfig() {
	err := config.LoadSource(
		env.NewSource([]string{"yggdrasil", "uuid_generator"}, []string{"uuid_generator"}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
