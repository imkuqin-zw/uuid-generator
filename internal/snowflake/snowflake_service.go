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

package snowflake

import (
	"context"

	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
)

type Service struct {
	generator *Snowflake
	api.UnimplementedSnowflakeServer
}

func NewService(generator *Snowflake) api.SnowflakeServer {
	return &Service{generator: generator}
}

func (s *Service) FetchNext(ctx context.Context, req *api.FetchSnowflakeNextReq) (*api.UUID, error) {
	uuid, err := s.generator.FetchNext()
	if err != nil {
		return nil, err
	}
	return &api.UUID{Val: uuid}, nil
}
