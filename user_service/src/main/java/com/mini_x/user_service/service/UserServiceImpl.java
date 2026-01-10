package com.mini_x.user_service.service;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.PrivateKey;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.Base64;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
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

@Service
public class UserServiceImpl implements UserService {
    private static final Logger logger = LoggerFactory.getLogger(UserServiceImpl.class);
    
    private final WriteRepo writeRepo;
    private final ReadRepo readRepo;
    private final UserCache userCache;
    private final TokenCacheService tokenCacheService;
    private final PrivateKey jwtPrivateKey;
    private final SecureRandom secureRandom;

    @Value("${jwt.expiration:604800}")
    private long jwtExpiration;
    
    @Value("${jwt.refresh.expiration:604800}")
    private long refreshTokenExpiration;
    
    private static final int HASH_ROUNDS = 12;
    private static final String HASH_ALGORITHM = "SHA-256";

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
        this.jwtPrivateKey = jwtPrivateKey;
        this.secureRandom = new SecureRandom();
    }

    @Override
    @Transactional("primaryTransactionManager")
    public void register(String username, String email, String password) {
        logger.info("Registering user: username={}, email={}", username, email);
        if (username == null || username.trim().isEmpty()) {
            logger.warn("Username is empty during registration");
            throw new InvalidInputException("Username cannot be empty");
        }
        if (email == null || email.trim().isEmpty()) {
            logger.warn("Email is empty during registration");
            throw new InvalidInputException("Email cannot be empty");
        }
        if (password == null || password.trim().isEmpty()) {
            logger.warn("Password is empty during registration");
            throw new InvalidInputException("Password cannot be empty");
        }
        try {
            Optional<User> existingUser = readRepo.findByEmail(email);
            if (existingUser.isPresent()) {
                logger.warn("User already exists for email: {}", email);
                throw new UserAlreadyExistsException(email);
            }
            String hashedPassword = hashPassword(password);
            User user = new User(username, email, hashedPassword);
            User savedUser = writeRepo.save(user);
            if (savedUser.getUserid() != null) {
                CachedUserData cacheData = new CachedUserData(savedUser.getUserid(), savedUser.getUsername());
                userCache.set(savedUser.getUserid(), cacheData);
                logger.info("User cached after registration: userId={}", savedUser.getUserid());
            }
            logger.info("User registered successfully: userId={}", savedUser.getUserid());
        } catch (Exception e) {
            logger.error("Error during user registration: {}", e.getMessage(), e);
            throw e;
        }
    }

    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public TokenPair login(String email, String password) {
        logger.info("Login attempt for email={}", email);
        if (email == null || email.trim().isEmpty()) {
            logger.warn("Email is empty during login");
            throw new InvalidInputException("Email cannot be empty");
        }
        if (password == null || password.trim().isEmpty()) {
            logger.warn("Password is empty during login");
            throw new InvalidInputException("Password cannot be empty");
        }
        try {
            Optional<User> userOpt = readRepo.findByEmail(email);
            if (userOpt.isEmpty()) {
                logger.warn("User not found for email: {}", email);
                throw new UserNotFoundException(email);
            }
            User user = userOpt.get();
            if (!verifyPassword(password, user.getPassword())) {
                logger.warn("Invalid credentials for email: {}", email);
                throw new InvalidCredentialsException();
            }
            logger.info("User login successful: userId={}", user.getUserid());
            return generateTokenPair(user.getUserid());
        } catch (Exception e) {
            logger.error("Error during login for email {}: {}", email, e.getMessage(), e);
            throw e;
        }
    }

    @Override
    public TokenPair refresh(String refreshToken) {
        logger.info("Refreshing token");
        if (refreshToken == null || refreshToken.trim().isEmpty()) {
            logger.warn("Refresh token is empty during refresh");
            throw new InvalidInputException("Refresh token cannot be empty");
        }
        try {
            String userId = tokenCacheService.getUserIdByRefreshToken(refreshToken);
            if (userId == null) {
                logger.warn("Invalid refresh token provided");
                throw new InvalidCredentialsException();
            }
            Long remainingTTL = tokenCacheService.getRefreshTokenTTL(refreshToken);
            String accessToken = Jwts.builder()
                .subject(userId)
                .issuer("users_service")
                .audience().add("api_gateway").and()
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + jwtExpiration * 1000))
                .signWith(jwtPrivateKey)
                .compact();
            String newRefreshToken = refreshToken;
            if (remainingTTL == null || remainingTTL > 0) {
                tokenCacheService.deleteRefreshToken(refreshToken);
                newRefreshToken = generateRefreshToken();
                tokenCacheService.storeRefreshToken(userId, newRefreshToken, refreshTokenExpiration);
                logger.info("Refresh token rotated for userId={}", userId);
            }
            logger.info("Token refreshed for userId={}", userId);
            return new TokenPair(accessToken, newRefreshToken);
        } catch (Exception e) {
            logger.error("Error during token refresh: {}", e.getMessage(), e);
            throw e;
        }
    }
    
    @Override
    public String logout(String refreshToken) {
        logger.info("Logout attempt");
        if (refreshToken == null || refreshToken.trim().isEmpty()) {
            logger.warn("Refresh token is empty during logout");
            throw new InvalidInputException("Refresh token cannot be empty");
        }
        try {
            boolean deleted = tokenCacheService.deleteRefreshToken(refreshToken);
            if (!deleted) {
                logger.warn("Invalid refresh token during logout");
                throw new InvalidCredentialsException();
            }
            logger.info("Logout successful");
            return "Logged out successfully";
        } catch (Exception e) {
            logger.error("Error during logout: {}", e.getMessage(), e);
            throw e;
        }
    }
    
    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public Map<String, List<String>> getUsersData(List<String> userIds) {
        logger.info("Fetching users data for userIds: {}", userIds);
        if (userIds == null || userIds.isEmpty()) {
            logger.warn("Empty userIds list in getUsersData");
            return createEmptyResponse();
        }
        List<String> usernames = new ArrayList<>();
        List<String> foundUserIds = new ArrayList<>();
        List<String> uncachedUserIds = new ArrayList<>();
        for (String userId : userIds) {
            Object cachedData = userCache.get(userId);
            if (cachedData != null && cachedData instanceof CachedUserData) {
                CachedUserData userData = (CachedUserData) cachedData;
                usernames.add(userData.getUsername());
                foundUserIds.add(userData.getUserId());
                logger.debug("Cache hit in getUsersData for userId={}", userId);
            } else {
                uncachedUserIds.add(userId);
                logger.debug("Cache miss in getUsersData for userId={}", userId);
            }
        }
        if (!uncachedUserIds.isEmpty()) {
            logger.info("Fetching from DB for userIds: {}", uncachedUserIds);
            List<User> users = readRepo.findUsersByIds(uncachedUserIds);
            for (User user : users) {
                usernames.add(user.getUsername());
                foundUserIds.add(user.getUserid());
                CachedUserData cacheData = new CachedUserData(user.getUserid(), user.getUsername());
                userCache.set(user.getUserid(), cacheData);
                logger.debug("User cached in getUsersData: userId={}", user.getUserid());
            }
        }
        Map<String, List<String>> response = new HashMap<>();
        response.put("usernames", usernames);
        response.put("userIds", foundUserIds);
        logger.info("getUsersData response: {}", response);
        return response;
    }
    
    private Map<String, List<String>> createEmptyResponse() {
        Map<String, List<String>> response = new HashMap<>();
        response.put("usernames", new ArrayList<>());
        response.put("userIds", new ArrayList<>());
        return response;
    }
    private TokenPair generateTokenPair(String userId) {
        logger.debug("Generating token pair for userId={}", userId);
        String accessToken = Jwts.builder()
            .subject(userId)
            .issuer("users_service")
            .audience().add("api_gateway").and()
            .issuedAt(new Date())
            .expiration(new Date(System.currentTimeMillis() + jwtExpiration * 1000))
            .signWith(jwtPrivateKey)
            .compact();
        String refreshToken = generateRefreshToken();
        tokenCacheService.storeRefreshToken(userId, refreshToken, refreshTokenExpiration);
        logger.debug("Token pair generated for userId={}", userId);
        return new TokenPair(accessToken, refreshToken);
    }
    
    private String generateRefreshToken() {
        byte[] randomBytes = new byte[32];
        secureRandom.nextBytes(randomBytes);
        String token = Base64.getUrlEncoder().withoutPadding().encodeToString(randomBytes);
        logger.debug("Generated refresh token");
        return token;
    }
    
    private String hashPassword(String password) {
        try {
            byte[] salt = new byte[16];
            secureRandom.nextBytes(salt);
            MessageDigest md = MessageDigest.getInstance(HASH_ALGORITHM);
            md.update(salt);
            byte[] hashedPassword = password.getBytes(StandardCharsets.UTF_8);
            for (int i = 0; i < HASH_ROUNDS; i++) {
                md.update(hashedPassword);
                hashedPassword = md.digest();
                md.reset();
            }
            String saltBase64 = Base64.getEncoder().encodeToString(salt);
            String hashBase64 = Base64.getEncoder().encodeToString(hashedPassword);
            logger.debug("Password hashed using {} with {} rounds", HASH_ALGORITHM, HASH_ROUNDS);
            return String.format("%s$%d$%s$%s", HASH_ALGORITHM, HASH_ROUNDS, saltBase64, hashBase64);
        } catch (NoSuchAlgorithmException e) {
            logger.error("Error hashing password: {}", e.getMessage(), e);
            throw new RuntimeException("Error hashing password", e);
        }
    }
    
    private boolean verifyPassword(String password, String storedHash) {
        try {
            String[] parts = storedHash.split("\\$");
            if (parts.length != 4) {
                logger.warn("Stored hash format invalid during password verification");
                return false;
            }
            String algorithm = parts[0];
            int iterations = Integer.parseInt(parts[1]);
            byte[] salt = Base64.getDecoder().decode(parts[2]);
            byte[] expectedHash = Base64.getDecoder().decode(parts[3]);
            MessageDigest md = MessageDigest.getInstance(algorithm);
            md.update(salt);
            byte[] hashedPassword = password.getBytes(StandardCharsets.UTF_8);
            for (int i = 0; i < iterations; i++) {
                md.update(hashedPassword);
                hashedPassword = md.digest();
                md.reset();
            }
            boolean match = MessageDigest.isEqual(hashedPassword, expectedHash);
            logger.debug("Password verification result: {}", match);
            return match;
        } catch (Exception e) {
            logger.error("Error verifying password: {}", e.getMessage(), e);
            return false;
        }
    }
}
