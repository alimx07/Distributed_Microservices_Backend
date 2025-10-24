package com.mini_x.user_service.cache;

import java.util.concurrent.TimeUnit;

import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;

@Service
public class TokenCacheService {
    
    private final RedisTemplate<String, Object> tokenRedisTemplate;
    private static final String REFRESH_TOKEN_PREFIX = "refresh_token:";
    
    public TokenCacheService(@Qualifier("tokenCacheRedisTemplate") RedisTemplate<String, Object> tokenRedisTemplate) {
        this.tokenRedisTemplate = tokenRedisTemplate;
    }
    

    public void storeRefreshToken(String userId, String refreshToken, long expirationSeconds) {
        String key = REFRESH_TOKEN_PREFIX + refreshToken;
        tokenRedisTemplate.opsForValue().set(key, userId, expirationSeconds, TimeUnit.SECONDS);
    }
    

    public String getUserIdByRefreshToken(String refreshToken) {
        String key = REFRESH_TOKEN_PREFIX + refreshToken;
        Object value = tokenRedisTemplate.opsForValue().get(key);
        
        if (value != null) {
            try {
                return value.toString();
            } catch (NumberFormatException e) {
                return null;
            }
        }
        return null;
    }
    
    public boolean deleteRefreshToken(String refreshToken) {
        String key = REFRESH_TOKEN_PREFIX + refreshToken;
        Boolean deleted = tokenRedisTemplate.delete(key);
        return deleted != null && deleted;
    }
    

    public boolean tokenExists(String refreshToken) {
        String key = REFRESH_TOKEN_PREFIX + refreshToken;
        Boolean exists = tokenRedisTemplate.hasKey(key);
        return exists != null && exists;
    }
    

    public boolean updateTokenExpiration(String refreshToken, long expirationSeconds) {
        String key = REFRESH_TOKEN_PREFIX + refreshToken;
        Boolean updated = tokenRedisTemplate.expire(key, expirationSeconds, TimeUnit.SECONDS);
        return updated != null && updated;
    }
}
