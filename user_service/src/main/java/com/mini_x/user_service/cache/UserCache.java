package com.mini_x.user_service.cache;

import java.time.Duration;

import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Component;

/**
 * Simple Redis cache service for user data
 * Uses standalone Redis instance
 * Cache TTL: 12 hours
 */
@Component
public class UserCache {
    
    private static final String USER_CACHE_PREFIX = "user:";
    private static final Duration CACHE_TTL = Duration.ofHours(12);
    
    private final RedisTemplate<String, Object> userRedisTemplate;
    
    public UserCache(@Qualifier("userCacheRedisTemplate") RedisTemplate<String, Object> userRedisTemplate) {
        this.userRedisTemplate = userRedisTemplate;
    }
    

    public Object get(String userId) {
        try {
            String key = USER_CACHE_PREFIX + userId;
            return userRedisTemplate.opsForValue().get(key);
        } catch (Exception e) {
           
            System.err.println("Cache get failed for userId " + userId + ": " + e.getMessage());
            return null;
        }
    }
    

    public void set(String userId, Object userData) {
        try {
            String key = USER_CACHE_PREFIX + userId;
            userRedisTemplate.opsForValue().set(key, userData, CACHE_TTL);
        } catch (Exception e) {
            
            System.err.println("Cache set failed for userId " + userId + ": " + e.getMessage());
        }
    }

    public void delete(Long userId) {
        try {
            String key = USER_CACHE_PREFIX + userId;
            userRedisTemplate.delete(key);
        } catch (Exception e) {
            
            System.err.println("Cache delete failed for userId " + userId + ": " + e.getMessage());
        }
    }
    

    public boolean exists(Long userId) {
        try {
            String key = USER_CACHE_PREFIX + userId;
            Boolean exists = userRedisTemplate.hasKey(key);
            return exists != null && exists;
        } catch (Exception e) {
           
            System.err.println("Cache exists check failed for userId " + userId + ": " + e.getMessage());
            return false;
        }
    }
}
