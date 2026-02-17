package com.mini_x.follow_service.exception;



public class FollowNotFoundException extends RuntimeException {
    public FollowNotFoundException(String followerId, String followingId) {
        super(String.format("User %s does not follow user %s", followerId, followingId));
    }
}
