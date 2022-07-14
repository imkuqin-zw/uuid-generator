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
