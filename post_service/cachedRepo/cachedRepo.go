package cachedRepo

import (
	"context"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
)

type CachedRepo interface {
	// Only posts will be cached for now
	CachePost(ctx context.Context, post models.CachedPost) error
	DeletePost(ctx context.Context, id int64) error
	GetPosts(ctx context.Context, ids []int64) ([]models.Post, error)
	UpdateLikesCounter(ctx context.Context, id int64)
	UpdateCommentsCounter(ctx context.Context, id int64)
	SyncCounters()
}
