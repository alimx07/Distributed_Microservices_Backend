package com.mini_x.follow_service.exception;



public class FollowNotFoundException extends RuntimeException {
    public FollowNotFoundException(Long followerId, Long followingId) {
        super(String.format("User %d does not follow user %d", followerId, followingId));
    }
}
