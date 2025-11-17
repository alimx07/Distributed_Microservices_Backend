package postRepo

import (
	"context"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
)

type PersistenceDB interface {
	CreatePost(ctx context.Context, post models.Post) (int64, error)
	DeletePost(ctx context.Context, id int64) error
	GetPosts(ctx context.Context, ids []int64) ([]models.Post, error)
	CreateComment(ctx context.Context, comment models.Comment) (int64, error)
	CreateLike(ctx context.Context, like models.Like) (int64, error)
	DeleteComment(ctx context.Context, id int64) error
	DeleteLike(ctx context.Context, post_id int64, userId int32) error
	GetComments(ctx context.Context, id int64) ([]models.Comment, error)
	GetLikes(ctx context.Context, id int64) ([]models.Like, error)
	GetCounters(ctx context.Context, ids []int64) ([]models.CachedCounter, error)
	UpdateCounters(ctx context.Context, counters []models.CachedCounter) error
}
