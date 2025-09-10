package postRepo

import (
	"context"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
)

type PostRepo interface {
	CreatePost(ctx context.Context, post models.Post) (int64, error)
	CreateComment(ctx context.Context, comment models.Comment) error
	CreateLike(ctx context.Context, like models.Like) error
	DeletePost(ctx context.Context, id int64, user_id int32) error
	DeleteComment(ctx context.Context, id int64, user_id int32) error
	DeleteLike(ctx context.Context, post_id int64, user_id int32) error
	GetPosts(ctx context.Context, ids []int64) ([]models.Post, error)
	GetComments(ctx context.Context, id int64) ([]models.Comment, error)
	GetLikes(ctx context.Context, id int64) ([]models.Like, error)
	GetCounters(ctx context.Context, ids []int64) ([]models.Post, error)
	UpdateCounters(ctx context.Context, counters []models.CachedCounter) error
}
