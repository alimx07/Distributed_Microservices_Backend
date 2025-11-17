package models

import "time"

type Config struct {
	// Primary (write) database
	DBHost     string
	DBPort     string
	DBUser     string
	DBName     string
	DBPassword string

	// Replica (read) database
	DBReplicaHost     string
	DBReplicaPort     string
	DBReplicaUser     string
	DBReplicaName     string
	DBReplicaPassword string

	CacheHost      string
	CachePort      string
	CachePassword  string
	ServerHost     string
	ServerPort     string
	ServerHttpPort string
}

type Post struct {
	CachedPost
	Likes_count    int64 `json:"likes_count"`
	Comments_count int64 `json:"comments_count"`
}

type CachedPost struct {
	Id         int64     `json:"id"`
	User_id    int32     `json:"user_id"`
	Content    string    `json:"content"`
	Created_at time.Time `json:"created_at"`
}

type CachedCounter struct {
	Id       int64
	Likes    int64
	Comments int64
}
type Comment struct {
	Id         int64
	User_id    int32
	Post_id    int64
	Content    string
	Created_at time.Time
}

type Like struct {
	Post_id int64
	User_id int32
}
