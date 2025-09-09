package models

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
	Id      int64
	User_id int32
	Content string
}

type Comment struct {
	Id      int64
	User_id int32
	Post_id int64
	Content string
}

type Like struct {
	Post_id int64
	User_id int32
}
