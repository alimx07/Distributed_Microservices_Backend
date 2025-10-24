package com.mini_x.user_service.cache;

import java.io.Serializable;


public class CachedUserData implements Serializable {
    
    private static final long serialVersionUID = 1L;
    
    private String userId;
    private String username;
    
    public CachedUserData() {
    }
    
    public CachedUserData(String userId, String username) {
        this.userId = userId;
        this.username = username;
    }
    
    public String getUserId() {
        return userId;
    }
    
    public void setUserId(String userId) {
        this.userId = userId;
    }
    
    public String getUsername() {
        return username;
    }
    
    public void setUsername(String username) {
        this.username = username;
    }
}
