package com.mini_x.user_service.service;

import java.security.PrivateKey;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.Base64;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.mini_x.user_service.cache.CachedUserData;
import com.mini_x.user_service.cache.TokenCacheService;
import com.mini_x.user_service.cache.UserCache;
import com.mini_x.user_service.dto.TokenPair;
import com.mini_x.user_service.entity.User;
import com.mini_x.user_service.exception.InvalidCredentialsException;
import com.mini_x.user_service.exception.InvalidInputException;
import com.mini_x.user_service.exception.UserAlreadyExistsException;
import com.mini_x.user_service.exception.UserNotFoundException;
import com.mini_x.user_service.repo.Read.ReadRepo;
import com.mini_x.user_service.repo.Write.WriteRepo;

import io.jsonwebtoken.Jwts;

/** 
 * Architecture Flow:
 * gRPC Service → User Service → Repository → Database
*/
@Service
public class UserServiceImpl implements UserService {
    
    private final WriteRepo writeRepo;
    private final ReadRepo readRepo;
    private final UserCache userCache;
    private final TokenCacheService tokenCacheService;
    private final BCryptPasswordEncoder passwordEncoder;
    private final PrivateKey jwtPrivateKey;
    private final SecureRandom secureRandom;

    @Value("${jwt.expiration:300}") // Default 5 minutes
    private long jwtExpiration;
    
    @Value("${jwt.refresh.expiration:604800}") // Default 7 days
    private long refreshTokenExpiration;

    // Components will be injected by spring boot
    public UserServiceImpl(
            WriteRepo writeRepo, 
            ReadRepo readRepo, 
            UserCache userCache,
            TokenCacheService tokenCacheService,
            PrivateKey jwtPrivateKey) {
        this.writeRepo = writeRepo;
        this.readRepo = readRepo;
        this.userCache = userCache;
        this.tokenCacheService = tokenCacheService;
        this.passwordEncoder = new BCryptPasswordEncoder();
        this.jwtPrivateKey = jwtPrivateKey;
        this.secureRandom = new SecureRandom();
    }

    @Override
    @Transactional("primaryTransactionManager")
    public TokenPair register(String username, String email, String password) {
        // Validate input
        if (username == null || username.trim().isEmpty()) {
            throw new InvalidInputException("Username cannot be empty");
        }
        if (email == null || email.trim().isEmpty()) {
            throw new InvalidInputException("Email cannot be empty");
        }
        if (password == null || password.trim().isEmpty()) {
            throw new InvalidInputException("Password cannot be empty");
        }
        
        
        Optional<User> existingUser = readRepo.findByEmail(email);
        if (existingUser.isPresent()) {
            throw new UserAlreadyExistsException(email);
        }
        
        
        String hashedPassword = passwordEncoder.encode(password);
        User user = new User(username, email, hashedPassword);
        User savedUser = writeRepo.save(user);
        
        // Cache the newly registered user
        if (savedUser.getUserid() != null) {
            CachedUserData cacheData = new CachedUserData(savedUser.getUserid(), savedUser.getUsername());
            userCache.set(savedUser.getUserid(), cacheData);
        }
        
        return generateTokenPair(savedUser.getUserid());
    }

    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public TokenPair login(String email, String password) {
        
        if (email == null || email.trim().isEmpty()) {
            throw new InvalidInputException("Email cannot be empty");
        }
        if (password == null || password.trim().isEmpty()) {
            throw new InvalidInputException("Password cannot be empty");
        }
        
        
        Optional<User> userOpt = readRepo.findByEmail(email);
        if (userOpt.isEmpty()) {
            throw new UserNotFoundException(email);
        }
        
        User user = userOpt.get();
        
        
        if (!passwordEncoder.matches(password, user.getPassword())) {
            throw new InvalidCredentialsException();
        }
        
        return generateTokenPair(user.getUserid());
    }

    @Override
    public TokenPair refresh(String refreshToken) {
        if (refreshToken == null || refreshToken.trim().isEmpty()) {
            throw new InvalidInputException("Refresh token cannot be empty");
        }
        
        // Retrieve user ID from Redis using refresh token
        String userId = tokenCacheService.getUserIdByRefreshToken(refreshToken);
        
        if (userId == null) {
            throw new InvalidCredentialsException();
        }
        
        // Check remaining TTL of refresh token
        Long remainingTTL = tokenCacheService.getRefreshTokenTTL(refreshToken);
        
        
        String accessToken = Jwts.builder()
            .subject(userId)
            .issuer("users_service")
            .audience().add("api_gateway").and()
            .issuedAt(new Date())
            .expiration(new Date(System.currentTimeMillis() + jwtExpiration * 1000))
            .signWith(jwtPrivateKey)
            .compact();
        
        // Keep the same refresh token if it is not expired
        // Otherwise, rotate it with a new one
        String newRefreshToken = refreshToken;
        
        if (remainingTTL == null || remainingTTL > 0) {
            // Delete old refresh token and generate new one
            tokenCacheService.deleteRefreshToken(refreshToken);
            newRefreshToken = generateRefreshToken();
            tokenCacheService.storeRefreshToken(userId, newRefreshToken, refreshTokenExpiration);
        }
        
        return new TokenPair(accessToken, newRefreshToken);
    }
    
    @Override
    public String logout(String refreshToken) {
        if (refreshToken == null || refreshToken.trim().isEmpty()) {
            throw new InvalidInputException("Refresh token cannot be empty");
        }
        
        // Delete refresh token from cache
        // Revoke
        boolean deleted = tokenCacheService.deleteRefreshToken(refreshToken);
        
        if (!deleted) {
            throw new InvalidCredentialsException();
        }
        
        return "Logged out successfully";
    }    
    
    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public Map<String, List<String>> getUsersData(List<String> userIds) {
        if (userIds == null || userIds.isEmpty()) {
            return createEmptyResponse();
        }
        
        List<String> usernames = new ArrayList<>();
        List<String> foundUserIds = new ArrayList<>();
        List<String> uncachedUserIds = new ArrayList<>();
        
        // Check cache first
        for (String userId : userIds) {
            Object cachedData = userCache.get(userId);
            if (cachedData != null && cachedData instanceof CachedUserData) {
                CachedUserData userData = (CachedUserData) cachedData;
                usernames.add(userData.getUsername());
                foundUserIds.add(userData.getUserId());
            } else {
                uncachedUserIds.add(userId);
            }
        }
        
        // Fetch uncached users from database
        if (!uncachedUserIds.isEmpty()) {
            List<User> users = readRepo.findUsersByIds(uncachedUserIds);
            
            for (User user : users) {
                usernames.add(user.getUsername());
                foundUserIds.add(user.getUserid());
                
                // Store in cache for next time
                CachedUserData cacheData = new CachedUserData(user.getUserid(), user.getUsername());
                userCache.set(user.getUserid(), cacheData);
            }
        }
        
        Map<String, List<String>> response = new HashMap<>();
        response.put("usernames", usernames);
        response.put("userIds", foundUserIds);
        
        return response;
    }
    
    private Map<String, List<String>> createEmptyResponse() {
        Map<String, List<String>> response = new HashMap<>();
        response.put("usernames", new ArrayList<>());
        response.put("userIds", new ArrayList<>());
        return response;
    }
    

    private TokenPair generateTokenPair(String userId) {
        // Generate access token (JWT) with RS256
        String accessToken = Jwts.builder()
            .subject(userId)
            .issuer("users_service")
            .audience().add("api_gateway").and()
            .issuedAt(new Date())
            .expiration(new Date(System.currentTimeMillis() + jwtExpiration * 1000))
            .signWith(jwtPrivateKey)  // RS256 is automatically selected for RSA private keys
            .compact();
        
        // Refresh Token : Opaque String
        String refreshToken = generateRefreshToken();
        
        
        tokenCacheService.storeRefreshToken(userId, refreshToken, refreshTokenExpiration);
        
        return new TokenPair(accessToken, refreshToken);
    }
    

    private String generateRefreshToken() {
        byte[] randomBytes = new byte[32];
        secureRandom.nextBytes(randomBytes);
        return Base64.getUrlEncoder().withoutPadding().encodeToString(randomBytes);
    }
}
