package data

import (
	"context"

	"github.com/imkuqin-zw/uuid-generator/internal/segment/domain"
	xgorm "github.com/imkuqin-zw/yggdrasil-gorm"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
	return xgorm.NewDB("uuid")
}

type SegmentRepo interface {
	domain.SegmentRepo
	GetSegmentByID(ctx context.Context, ID string) (*domain.Segment, error)
}
