package postRepo

import (
	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
)

type PostRepo interface {
	CreatePost(post models.Post) error
	CreateComment(comment models.Comment) error
	CreateLike(like models.Like) error
	DeletePost(post models.Post) error
	DeleteComment(comment models.Comment) error
	DeleteLike(like models.Like) error
}
