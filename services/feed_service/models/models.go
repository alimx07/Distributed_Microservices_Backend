package models

type KafkaConfig struct {
	BootStrapServers string
	GroupID          string
	OffsetReset      string
	FetchMinBytes    string
	Topics           []string
}

type RedisConfig struct {
	ClusterAddr []string
	Password    string
}

type ServerConfig struct {
	ServerHost     string
	ServerPort     string
	ServerHTTPPort string
	PostService    string
	UserService    string
	FollowService  string
	// EtcdEndpoints  string
	HostName string
}

type FeedItem struct {
	PostId     string `json:"post_id"`
	UserId     string `json:"user_id"`
	Created_at int64  `json:"created_at"`
}

type Cursor struct {
	UserId   string
	Cursor   string
	PageSize int32
}

type Post struct {
	UserName       string
	UserID         string
	Content        string
	Created_at     int64
	Likes_count    int64
	Comments_count int64
}
