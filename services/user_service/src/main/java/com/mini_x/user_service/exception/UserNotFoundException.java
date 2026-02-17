package com.mini_x.user_service.exception;

public class UserNotFoundException extends RuntimeException {
    public UserNotFoundException(String email) {
        super("User with email " + email + " not found");
    }
    
    public UserNotFoundException(Long userId) {
        super("User with ID " + userId + " not found");
    }
}
