package postRepo

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
	pb "github.com/alimx07/Distributed_Microservices_Backend/services_bindings_go"
	"google.golang.org/protobuf/proto"
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
	var postId string
	tx.QueryRowContext(ctx,
		`INSERT INTO posts (user_id, content) 
		VALUES ($1, $2) RETURNING  id`,
		post.User_id, post.Content).Scan(&postId)
	if err != nil {
		tx.Rollback()
		log.Println("Error creating post: ", err.Error())
		return err
	}
	evt := &pb.KPostCreated{
		Id:        postId,
		UserId:    post.User_id,
		Content:   post.Content,
		CreatedAt: time.Now().UnixMilli(),
	}
	payload, _ := proto.Marshal(evt)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox (topic ,kafka_key, payload) VALUES ($1 , $2, $3)`,
		"posts.created", post.User_id, payload)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert Post in outbox DB: ", err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) CreateComment(ctx context.Context, comment models.Comment) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	var commentID int64
	tx.QueryRowContext(ctx,
		`INSERT INTO comments (user_id, post_id ,content) 
		VALUES ($1, $2 , $3) RETURNING  id`,
		comment.User_id, comment.Post_id, comment.Content).Scan(&commentID)
	if err != nil {
		tx.Rollback()
		log.Println("Error creating Coment: ", err.Error())
		return err
	}
	evt := &pb.KCommentCreated{
		Id:        commentID,
		UserId:    comment.User_id,
		PostId:    comment.Post_id,
		Content:   comment.Content,
		CreatedAt: time.Now().UnixMilli(),
	}
	payload, _ := proto.Marshal(evt)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox (topic ,kafka_key, payload) VALUES ($1 , $2, $3)`,
		"comments.created", comment.User_id, payload)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert Commnet in outbox DB: ", err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) CreateLike(ctx context.Context, like models.Like) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	var likeID int64
	tx.QueryRowContext(ctx,
		`INSERT INTO likes (user_id, post_id) 
		VALUES ($1, $2) RETURNING  id`,
		like.User_id, like.Post_id).Scan(&likeID)
	if err != nil {
		tx.Rollback()
		log.Println("Error creating Like: ", err.Error())
		return err
	}
	evt := &pb.KLikeCreated{
		UserId:    like.User_id,
		PostId:    like.Post_id,
		CreatedAt: time.Now().UnixMilli(),
	}
	payload, _ := proto.Marshal(evt)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox (topic ,kafka_key, payload) VALUES ($1 , $2 , $3)`,
		"likes.created", like.User_id, payload)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert Commnet in outbox DB: ", err.Error())
		return err
	}
	return nil
}

func (ps *PostgresRepo) DeletePost(ctx context.Context, post models.Post) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	tx.ExecContext(ctx,
		`DELETE FROM posts where id = $1`, post.Id)
	if err != nil {
		tx.Rollback()
		log.Printf("Error Deleting post{%v} : %v\n", post.Id, err.Error())
		return err
	}
	evt := &pb.KPostDeleted{
		Id: post.Id,
	}
	payload, _ := proto.Marshal(evt)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox (topic ,kafka_key, payload) VALUES ($1 , $2, $3)`,
		"posts.deleted", post.User_id, payload)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert Post Deletion in outbox DB: ", err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) DeleteComment(ctx context.Context, comment models.Comment) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	tx.ExecContext(ctx,
		`DELETE FROM comments where id = $1`, comment.Id)
	if err != nil {
		tx.Rollback()
		log.Printf("Error Deleting comment{%v} : %v\n", comment.Id, err.Error())
		return err
	}
	evt := &pb.KCommentDeleted{
		Id: comment.Id,
	}
	payload, _ := proto.Marshal(evt)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox (topic ,kafka_key, payload) VALUES ($1 , $2, $3)`,
		"comments.deleted", comment.User_id, payload)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert Comment Deletion in outbox DB: ", err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) DeleteLike(ctx context.Context, like models.Like) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	tx.ExecContext(ctx,
		`DELETE FROM likes where user_id = $1 AND post_id = $2`, like.User_id, like.Post_id)
	if err != nil {
		tx.Rollback()
		log.Printf("Error Deleting of user{%v} for post{%v} : %v\n", like.User_id, like.Post_id, err.Error())
		return err
	}
	evt := &pb.KLikeDeleted{
		UserId: like.User_id,
		PostId: like.Post_id,
	}
	payload, _ := proto.Marshal(evt)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox (topic ,kafka_key, payload) VALUES ($1 , $2, $3)`,
		"posts.deleted", like.User_id, payload)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert Like Deletion in outbox DB: ", err.Error())
		return err
	}
	return nil
}
