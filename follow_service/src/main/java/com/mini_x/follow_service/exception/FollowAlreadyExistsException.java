package com.mini_x.follow_service.exception;



public class FollowAlreadyExistsException extends RuntimeException{
    public FollowAlreadyExistsException(Long followerId , Long followeeId) {
        super(String.format("User %d already follows User %d" , followerId , followeeId));
    }
}