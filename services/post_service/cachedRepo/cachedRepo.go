package cachedRepo

import (
	"context"

	"github.com/alimx07/Distributed_Microservices_Backend/services/post_service/models"
)

type CachedRepo interface {
	// Only posts will be cached for now
	CachePost(ctx context.Context, post models.CachedPost) error
	DeletePost(ctx context.Context, id string) error
	GetPosts(ctx context.Context, ids []string) ([]models.Post, error)
	UpdateLikesCounter(ctx context.Context, id string, delta int64)
	UpdateCommentsCounter(ctx context.Context, id string, delta int64)
	SyncCounters()
	Close()
}
