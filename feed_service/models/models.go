package models

type KafkaConfig struct {
	BootStrapServers string
	GroupID          string
	OffsetReset      string
	FetchMinBytes    string
	Topics           []string
}

type RedisConfig struct {
	Addr     string
	Port     string
	Password string
}

type ServerConfig struct {
	ServerHost    string
	ServerPort    string
	PostService   string
	UserService   string
	FollowService string
}
type FeedItem struct {
	PostId     int64 `json:"post_id"`
	UserId     int64 `json:"user_id"`
	Created_at int64 `json:"created_at"`
}

type Cursor struct {
	UserId   int64
	Cursor   string
	PageSize int32
}

type Post struct {
	UserName       string
	UserID         int32
	Content        string
	Created_at     int64
	Likes_count    int64
	Comments_count int64
}
