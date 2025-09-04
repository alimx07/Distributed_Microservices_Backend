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
	Id             int32
	User_id        string
	Content        string
	Likes_count    int32
	Comments_count int32
}

type Comment struct {
	Id      int64
	User_id int32
	Post_id int32
	Content string
}

type Like struct {
	Post_id int32
	User_id int32
}

// another configs will be added later
type KafkaConfig struct {
	Brokers         []string
	InputTopic      string
	BatchSize       string
	ComperssionType string
	Acks            string
	MaxInFlight     string
	Idempotence     string
}
