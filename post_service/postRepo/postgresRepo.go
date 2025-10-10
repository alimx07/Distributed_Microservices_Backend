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
	primaryDB *sql.DB // For writes
	replicaDB *sql.DB // For reads
}

func NewPostgresRepo(primaryDB, replicaDB *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		primaryDB: primaryDB,
		replicaDB: replicaDB,
	}
}

// Write operations use primaryDB
func (ps *PostgresRepo) CreatePost(ctx context.Context, post models.Post) (int64, error) {
	var id int64
	err := ps.primaryDB.QueryRowContext(ctx,
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
	_, err := ps.primaryDB.ExecContext(ctx,
		`INSERT INTO comments (user_id, post_id ,content) 
        VALUES ($1, $2 , $3)`,
		comment.User_id, comment.Post_id, comment.Content)
	if err != nil {
		log.Println("Error creating Comment: ", err.Error())
		return err
	}
	return nil
}

func (ps *PostgresRepo) CreateLike(ctx context.Context, like models.Like) error {
	_, err := ps.primaryDB.ExecContext(ctx,
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
	_, err := ps.primaryDB.ExecContext(ctx,
		`DELETE FROM posts where id = $1`, id)
	if err != nil {
		log.Printf("Error Deleting post{%v} : %v\n", id, err.Error())
		return err
	}
	return nil
}

func (ps *PostgresRepo) DeleteComment(ctx context.Context, id int64) error {
	_, err := ps.primaryDB.ExecContext(ctx,
		`DELETE FROM comments where id = $1`, id)
	if err != nil {
		log.Printf("Error Deleting comment{%v} : %v\n", id, err.Error())
		return err
	}
	return nil
}

func (ps *PostgresRepo) DeleteLike(ctx context.Context, post_id int64, userId int32) error {
	_, err := ps.primaryDB.ExecContext(ctx,
		`DELETE FROM likes where post_id = $1 AND user_id = $2`, post_id, userId)
	if err != nil {
		log.Printf("Error Deleting like{%v} : %v\n", post_id, err.Error())
		return err
	}
	return nil
}

// Read operations use replicaDB
func (ps *PostgresRepo) GetPosts(ctx context.Context, ids []int64) ([]models.Post, error) {
	if len(ids) == 0 {
		return []models.Post{}, nil
	}

	rows, err := ps.replicaDB.QueryContext(ctx,
		`SELECT id, user_id, content , created_at , likes_count , comments_count FROM posts WHERE id = ANY($1)`,
		pq.Array(ids))
	if err != nil {
		log.Println("Error querying posts: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	posts := make([]models.Post, 0, len(ids))
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
	rows, err := ps.replicaDB.QueryContext(ctx,
		`SELECT id, user_id, post_id, content, created_at FROM comments WHERE post_id = $1`, id)
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
	return comments, nil
}

func (ps *PostgresRepo) GetLikes(ctx context.Context, id int64) ([]models.Like, error) {
	rows, err := ps.replicaDB.QueryContext(ctx,
		`SELECT user_id, post_id FROM likes WHERE post_id = $1`, id)
	if err != nil {
		log.Println("Error querying likes: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	var likes []models.Like
	for rows.Next() {
		var like models.Like
		if err := rows.Scan(&like.User_id, &like.Post_id); err != nil {
			log.Println("Error scanning like row: ", err.Error())
			return nil, err
		}
		likes = append(likes, like)
	}
	return likes, nil
}

func (ps *PostgresRepo) GetCounters(ctx context.Context, ids []int64) ([]models.CachedCounter, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := ps.replicaDB.QueryContext(ctx,
		`SELECT id, likes_count , comments_count FROM posts WHERE id = ANY($1)`,
		pq.Array(ids))
	if err != nil {
		log.Println("Error querying posts: ", err.Error())
		return nil, err
	}
	defer rows.Close()

	cnts := make([]models.CachedCounter, 0, len(ids))
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

// Write operation - uses primaryDB
func (ps *PostgresRepo) UpdateCounters(ctx context.Context, counters []models.CachedCounter) error {
	values := make([]string, 0, len(counters))
	for _, cnt := range counters {
		values = append(values, fmt.Sprintf("(%d,%d,%d)", cnt.Id, cnt.Likes, cnt.Comments))
	}
	query := fmt.Sprintf(`UPDATE posts p SET 
                        likes_count = p.likes_count + v.likes, 
                        comments_count = p.comments_count + v.comments
                        FROM (VALUES %s) AS v(id, likes, comments) 
                        WHERE v.id = p.id`, strings.Join(values, ","))
	_, err := ps.primaryDB.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error In updating Counters: %v", err.Error())
		return err
	}
	return nil
}
