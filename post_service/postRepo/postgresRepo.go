package postRepo

import (
	"context"
	"database/sql"
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
func (ps *PostgresRepo) CreatePost(ctx context.Context, post models.Post) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		log.Println("Error in starting the transactions: ", err.Error())
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO posts (user_id, content) 
		VALUES ($1, $2)`,
		post.User_id, post.Content)
	if err != nil {
		tx.Rollback()
		log.Println("Error creating post: ", err.Error())
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Println("Failed to Commit the Transaction: ", err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) CreateComment(ctx context.Context, comment models.Comment) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		log.Println("Error in starting the transactions: ", err.Error())
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO comments (user_id, post_id ,content) 
		VALUES ($1, $2 , $3)`,
		comment.User_id, comment.Post_id, comment.Content)
	if err != nil {
		tx.Rollback()
		log.Println("Error creating Coment: ", err.Error())
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Println("Failed to Commit the Transaction: ", err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) CreateLike(ctx context.Context, like models.Like) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		log.Println("Error in starting the transactions: ", err.Error())
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO likes (user_id, post_id) 
		VALUES ($1, $2)`,
		like.User_id, like.Post_id)
	if err != nil {
		tx.Rollback()
		log.Println("Error creating Like: ", err.Error())
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Println("Failed to Commit the Transaction: ", err.Error())
		return err
	}
	return nil
}

func (ps *PostgresRepo) DeletePost(ctx context.Context, id int64, user_id int32) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		log.Println("Error in starting the transactions: ", err.Error())
		return err
	}
	_, err = tx.ExecContext(ctx,
		`DELETE FROM posts where id = $1`, id)
	if err != nil {
		tx.Rollback()
		log.Printf("Error Deleting post{%v} : %v\n", id, err.Error())
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Println("Failed to Commit the Transaction: ", err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) DeleteComment(ctx context.Context, id int64, user_id int32) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		log.Println("Error in starting the transactions: ", err.Error())
		return err
	}
	_, err = tx.ExecContext(ctx,
		`DELETE FROM comments where id = $1`, id)
	if err != nil {
		tx.Rollback()
		log.Printf("Error Deleting comment{%v} : %v\n", id, err.Error())
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Println("Failed to Commit the Transaction: ", err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) DeleteLike(ctx context.Context, post_id int64, user_id int32) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		log.Println("Error in starting the transactions: ", err.Error())
		return err
	}
	_, err = tx.ExecContext(ctx,
		`DELETE FROM likes where user_id = $1 AND post_id = $2`, user_id, post_id)
	if err != nil {
		tx.Rollback()
		log.Printf("Error Deleting of user{%v} for post{%v} : %v\n", user_id, post_id, err.Error())
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Println("Failed to Commit the Transaction: ", err.Error())
		return err
	}
	return nil
}
