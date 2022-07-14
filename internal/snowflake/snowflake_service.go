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
