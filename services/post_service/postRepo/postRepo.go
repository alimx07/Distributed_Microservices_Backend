package postRepo

import (
	"context"

	"github.com/alimx07/Distributed_Microservices_Backend/services/post_service/models"
)

type PersistenceDB interface {
	CreatePost(ctx context.Context, post models.Post) (string, error)
	DeletePost(ctx context.Context, id string) error
	GetPosts(ctx context.Context, ids []string) ([]models.Post, error)
	CreateComment(ctx context.Context, comment models.Comment) (string, error)
	CreateLike(ctx context.Context, like models.Like) error
	DeleteComment(ctx context.Context, id string) error
	DeleteLike(ctx context.Context, post_id string, userId string) error
	GetComments(ctx context.Context, id string) ([]models.Comment, error)
	GetLikes(ctx context.Context, id string) ([]models.Like, error)
	GetCounters(ctx context.Context, ids []string) ([]models.CachedCounter, error)
	UpdateCounters(ctx context.Context, counters []models.CachedCounter) error
	Close()
}
