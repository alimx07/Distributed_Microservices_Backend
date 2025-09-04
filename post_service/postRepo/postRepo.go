package postRepo

import (
	"context"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
)

type PostRepo interface {
	CreatePost(ctx context.Context, post models.Post) error
	CreateComment(ctx context.Context, comment models.Comment) error
	CreateLike(ctx context.Context, like models.Like) error
	DeletePost(ctx context.Context, post models.Post) error
	DeleteComment(ctx context.Context, comment models.Comment) error
	DeleteLike(ctx context.Context, like models.Like) error
}
