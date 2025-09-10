package cachedRepo

import (
	"context"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/postRepo"
)

type CachedRepo interface {
	postRepo.PostRepo

	UpdateLikesCounter(ctx context.Context, id int64)
	UpdateCommentsCounter(ctx context.Context, id int64)
}
