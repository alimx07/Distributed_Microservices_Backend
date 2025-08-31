package models

import "time"

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBName     string
	DBPassword string
	ServerHost string
	ServerPort string
}
type Post struct {
	Id             int
	User_id        string
	Content        string
	Likes_count    int
	Comments_count int
	Created_at     time.Duration
}

type Comment struct {
	Id         int64
	User_id    int
	Post_id    int
	Content    string
	Created_at time.Duration
}

type Like struct {
	Post_id int
	User_id int
}
