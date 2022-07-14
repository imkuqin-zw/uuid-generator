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
