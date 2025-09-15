package models

type KafkaConfig struct {
	BootStrapServers string
	GroupID          string
	OffsetReset      string
	FetchMinBytes    string
}

type RedisConfig struct {
	Addr     string
	Port     string
	Password string
}

type ServerConfig struct {
	ServerHost string
	ServerPort string
}
type FeedItem struct {
	PostId     int64 `json:"post_id"`
	UserId     int64 `json:"user_id"`
	Created_at int64 `json:"created_at"`
}

type Cursor struct {
	UserId int64
	Cursor string
}
