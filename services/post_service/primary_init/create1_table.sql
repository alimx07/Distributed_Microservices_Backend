CREATE TABLE IF NOT EXISTS posts(
	post_id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	likes_count BIGINT DEFAULT 0,
	comments_count BIGINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS comments (
	comment_id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	post_id TEXT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
	content TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS likes(
	post_id TEXT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
	user_id TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	PRIMARY KEY (post_id , user_id)
);

CREATE INDEX idx_user_posts ON posts(user_id , created_at);
CREATE INDEX idx_post_commnets ON comments(post_id , created_at);
CREATE INDEX idx_post_likes ON likes(post_id, created_at);