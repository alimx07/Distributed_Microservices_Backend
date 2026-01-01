CREATE TABLE IF NOT EXISTS follow (
    follower_id VARCHAR(26) NOT NULL,
    followee_id VARCHAR(26) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id , followee_id)
);

CREATE INDEX idx_follows_follower_id ON follow(follower_id);
CREATE INDEX idx_follows_followee_id ON follow(followee_id);
