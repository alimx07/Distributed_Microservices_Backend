package com.mini_x.follow_service.exception;



public class FollowAlreadyExistsException extends RuntimeException{
    public FollowAlreadyExistsException(String followerId , String followeeId) {
        super(String.format("User %s already follows User %s" , followerId , followeeId));
    }
}