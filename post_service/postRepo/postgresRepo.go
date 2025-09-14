package postRepo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
	"github.com/lib/pq"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}
func (ps *PostgresRepo) CreatePost(ctx context.Context, post models.Post) (int64, error) {
	var id int64
	err := ps.db.QueryRowContext(ctx,
		`INSERT INTO posts (user_id, content) 
		VALUES ($1, $2) RETURNING id`,
		post.User_id, post.Content).Scan(&id)
	if err != nil {
		log.Println("Error creating post: ", err.Error())
		return 0, err
	}
	return id, nil
}
func (ps *PostgresRepo) CreateComment(ctx context.Context, comment models.Comment) error {
	_, err := ps.db.ExecContext(ctx,
		`INSERT INTO comments (user_id, post_id ,content) 
		VALUES ($1, $2 , $3)`,
		comment.User_id, comment.Post_id, comment.Content)
	if err != nil {
		log.Println("Error creating Coment: ", err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) CreateLike(ctx context.Context, like models.Like) error {
	_, err := ps.db.ExecContext(ctx,
		`INSERT INTO likes (user_id, post_id) 
		VALUES ($1, $2)`,
		like.User_id, like.Post_id)
	if err != nil {
		log.Println("Error creating Like: ", err.Error())
		return err
	}
	return nil
}

func (ps *PostgresRepo) DeletePost(ctx context.Context, id int64) error {
	_, err := ps.db.ExecContext(ctx,
		`DELETE FROM posts where id = $1`, id)
	if err != nil {
		log.Printf("Error Deleting post{%v} : %v\n", id, err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) DeleteComment(ctx context.Context, id int64) error {
	_, err := ps.db.ExecContext(ctx,
		`DELETE FROM comments where id = $1`, id)
	if err != nil {
		log.Printf("Error Deleting comment{%v} : %v\n", id, err.Error())
		return err
	}
	return nil
}
func (ps *PostgresRepo) DeleteLike(ctx context.Context, post_id int64, userId int32) error {
	_, err := ps.db.ExecContext(ctx,
		`DELETE FROM likes where user_id = $1 AND post_id = $2`, userId, post_id)
	if err != nil {
		log.Printf("Error Deleting of user{%v} for post{%v} : %v\n", userId, post_id, err.Error())
		return err
	}
	return nil
}

func (ps *PostgresRepo) GetPosts(ctx context.Context, ids []int64) ([]models.Post, error) {
	if len(ids) == 0 {
		return []models.Post{}, nil
	}

	rows, err := ps.db.QueryContext(ctx,
		`SELECT id, user_id, content , created_at , likes_count , comments_count FROM posts WHERE id = ANY($1)`,
		pq.Array(ids))
	if err != nil {
		log.Println("Error querying posts: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	posts := make([]models.Post, len(ids))
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.Id, &post.User_id, &post.Content, &post.Created_at, &post.Likes_count, &post.Comments_count); err != nil {
			log.Println("Error scanning post row: ", err.Error())
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating post rows: ", err.Error())
		return nil, err
	}

	return posts, nil
}

func (ps *PostgresRepo) GetComments(ctx context.Context, id int64) ([]models.Comment, error) {

	rows, err := ps.db.QueryContext(ctx,
		`SELECT id, user_id, post_id, content, created_at 
        FROM comments 
        WHERE post_id = $1
        ORDER BY created_at DESC`,
		id)
	if err != nil {
		log.Println("Error querying comments: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.Id, &comment.User_id, &comment.Post_id, &comment.Content, &comment.Created_at); err != nil {
			log.Println("Error scanning comment row: ", err.Error())
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating comment rows: ", err.Error())
		return nil, err
	}

	return comments, nil
}

func (ps *PostgresRepo) GetLikes(ctx context.Context, id int64) ([]models.Like, error) {

	rows, err := ps.db.QueryContext(ctx,
		`SELECT post_id, user_id, created_at 
        FROM likes 
        WHERE post_id = $1`,
		id)
	if err != nil {
		log.Println("Error querying likes: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	var likes []models.Like
	for rows.Next() {
		var like models.Like
		if err := rows.Scan(&like.Post_id, &like.User_id); err != nil {
			log.Println("Error scanning like row: ", err.Error())
			return nil, err
		}
		likes = append(likes, like)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating like rows: ", err.Error())
		return nil, err
	}

	return likes, nil
}

func (ps *PostgresRepo) GetCounters(ctx context.Context, ids []int64) ([]models.CachedCounter, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := ps.db.QueryContext(ctx,
		`SELECT id, likes_count , comments_count FROM posts WHERE id = ANY($1)`,
		pq.Array(ids))
	if err != nil {
		log.Println("Error querying posts: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	cnts := make([]models.CachedCounter, len(ids))
	for rows.Next() {
		var cnt models.CachedCounter
		if err := rows.Scan(&cnt.Id, &cnt.Likes, &cnt.Comments); err != nil {
			log.Println("Error scanning post row: ", err.Error())
			return nil, err
		}
		cnts = append(cnts, cnt)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating post rows: ", err.Error())
		return nil, err
	}

	return cnts, nil
}

func (ps *PostgresRepo) UpdateCounters(ctx context.Context, counters []models.CachedCounter) error {
	values := make([]string, 0, len(counters))
	for _, cnt := range counters {
		values = append(values, fmt.Sprintf("%d,%d,%d", cnt.Id, cnt.Likes, cnt.Comments))
	}
	query := fmt.Sprintf(`UPDATE posts p SET 
						likes = p.likes + v.likes, comments = p.comments + v.comments
						FROM (VALUES %s)AS v(id , likes , comments) 
						WHERE v.id = p.id`, strings.Join(values, ","))
	_, err := ps.db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error In updating Counters: %v", err.Error())
		return err
	}
	return nil
}
