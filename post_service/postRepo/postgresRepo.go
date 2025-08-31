package postRepo

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}
func (ps *PostgresRepo) CreatePost(post models.Post) error {
	query := `INSERT INTO posts (user_id, content) 
              VALUES ($1, $2)`
	_, err := ps.db.Exec(query, post.User_id, post.Content)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		return fmt.Errorf("failed to create post: %w", err)
	}

	return nil
}
func (ps *PostgresRepo) CreateComment(comment models.Comment) error {
	query := `INSERT INTO comments (post_id, user_id, content) 
              VALUES ($1, $2, $3)`

	_, err := ps.db.Exec(query, comment.Post_id, comment.User_id, comment.Content)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}
func (ps *PostgresRepo) CreateLike(like models.Like) error {
	query := `INSERT INTO likes (post_id, user_id) 
              VALUES ($1, $2)`

	_, err := ps.db.Exec(query, like.Post_id, like.User_id)
	if err != nil {
		log.Printf("Error creating like: %v", err)
		return fmt.Errorf("failed to create like: %w", err)
	}

	return nil
}

func (ps *PostgresRepo) DeletePost(post models.Post) error {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := ps.db.Exec(query, post.Id)
	if err != nil {
		log.Printf("Error deleting post: %v", err)
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("post with ID %v not found", post.Id)
	}

	return nil
}
func (ps *PostgresRepo) DeleteComment(comment models.Comment) error {
	query := `DELETE FROM comments WHERE id = $1`

	result, err := ps.db.Exec(query, comment.Id)
	if err != nil {
		log.Printf("Error deleting comment: %v", err)
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("comment with ID %v not found", comment.Id)
	}

	return nil
}
func (ps *PostgresRepo) DeleteLike(like models.Like) error {
	query := `DELETE FROM likes WHERE post_id = $1 AND user_id= $2`

	result, err := ps.db.Exec(query, like.Post_id, like.User_id)
	if err != nil {
		log.Printf("Error deleting like: %v", err)
		return fmt.Errorf("failed to delete like: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("like of user {%v} for post {%v} not found", like.User_id, like.Post_id)
	}

	return nil
}
