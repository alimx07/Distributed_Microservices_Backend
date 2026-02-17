package com.mini_x.user_service.cache;

import java.time.Duration;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Component;

@Component
public class UserCache {
    private static final Logger logger = LoggerFactory.getLogger(UserCache.class);
    
    private static final String USER_CACHE_PREFIX = "user:";
    private static final Duration CACHE_TTL = Duration.ofHours(12);
    
    private final RedisTemplate<String, Object> userRedisTemplate;
    
    public UserCache(@Qualifier("userCacheRedisTemplate") RedisTemplate<String, Object> userRedisTemplate) {
        this.userRedisTemplate = userRedisTemplate;
    }
    

    public Object get(String userId) {
        try {
            String key = USER_CACHE_PREFIX + userId;
            logger.debug("Getting user from cache: {}", key);
            Object value = userRedisTemplate.opsForValue().get(key);
            if (value != null) {
                logger.info("Cache hit for userId: {}", userId);
            } else {
                logger.info("Cache miss for userId: {}", userId);
            }
            return value;
        } catch (Exception e) {
            logger.error("Cache get failed for userId {}: {}", userId, e.getMessage(), e);
            return null;
        }
    }
    

    public void set(String userId, Object userData) {
        try {
            String key = USER_CACHE_PREFIX + userId;
            logger.debug("Setting user in cache: {}", key);
            userRedisTemplate.opsForValue().set(key, userData, CACHE_TTL);
            logger.info("User cached for userId: {}", userId);
        } catch (Exception e) {
            logger.error("Cache set failed for userId {}: {}", userId, e.getMessage(), e);
        }
    }

    public void delete(Long userId) {
        try {
            String key = USER_CACHE_PREFIX + userId;
            logger.debug("Deleting user from cache: {}", key);
            userRedisTemplate.delete(key);
            logger.info("User cache deleted for userId: {}", userId);
        } catch (Exception e) {
            logger.error("Cache delete failed for userId {}: {}", userId, e.getMessage(), e);
        }
    }
    

    public boolean exists(Long userId) {
        try {
            String key = USER_CACHE_PREFIX + userId;
            logger.debug("Checking if user exists in cache: {}", key);
            Boolean exists = userRedisTemplate.hasKey(key);
            logger.info("Cache exists for userId {}: {}", userId, exists);
            return exists != null && exists;
        } catch (Exception e) {
            logger.error("Cache exists check failed for userId {}: {}", userId, e.getMessage(), e);
            return false;
        }
    }
}
