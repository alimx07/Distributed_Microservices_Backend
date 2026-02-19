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

	CacheAddrs     []string
	CachePassword  string
	ServerHost     string
	ServerPort     string
	ServerHttpPort string

	// EtcdEndpoints string
	HostName string
}

type Post struct {
	CachedPost
	Likes_count    int64 `json:"likes_count"`
	Comments_count int64 `json:"comments_count"`
}

type CachedPost struct {
	Id         string    `json:"id"`
	User_id    string    `json:"user_id"`
	Content    string    `json:"content"`
	Created_at time.Time `json:"created_at"`
}

type CachedCounter struct {
	Id       string
	Likes    int64
	Comments int64
}
type Comment struct {
	Id         string
	User_id    string
	Post_id    string
	Content    string
	Created_at time.Time
}

type Like struct {
	Post_id string
	User_id string
}
