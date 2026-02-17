package com.mini_x.user_service.grpc;

import java.util.List;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import com.mini_x.user_service.dto.TokenPair;
import com.mini_x.user_service.exception.InvalidCredentialsException;
import com.mini_x.user_service.exception.InvalidInputException;
import com.mini_x.user_service.exception.UserAlreadyExistsException;
import com.mini_x.user_service.exception.UserNotFoundException;
import com.mini_x.user_service.service.UserService;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import net.devh.boot.grpc.server.service.GrpcService;

@Component
@GrpcService
public class UserServiceGrpcImpl extends UserServiceGrpc.UserServiceImplBase {
    private static final Logger logger = LoggerFactory.getLogger(UserServiceGrpcImpl.class);

    private final UserService userService;

    public UserServiceGrpcImpl(UserService userService) {
        this.userService = userService;
    }

    @Override
    public void register(RegisterRequest request, StreamObserver<RegisterResponse> responseObserver) {
        logger.info("gRPC register called: username={}, email={}", request.getUsername(), request.getEmail());
        try {
            userService.register(
                request.getUsername(),
                request.getEmail(),
                request.getPassword()
            );
            RegisterResponse response = RegisterResponse.newBuilder()
                .setMessage("User Registered Successfully")
                .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            logger.info("gRPC register successful for email={}", request.getEmail());
        } catch (InvalidInputException e) {
            logger.warn("gRPC register invalid input: {}", e.getMessage());
            responseObserver.onError(Status.INVALID_ARGUMENT
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (UserAlreadyExistsException e) {
            logger.warn("gRPC register user already exists: {}", e.getMessage());
            responseObserver.onError(Status.ALREADY_EXISTS
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (Exception e) {
            logger.error("gRPC register internal error: {}", e.getMessage(), e);
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }

    @Override
    public void login(LoginRequest request, StreamObserver<TokenResponse> responseObserver) {
        logger.info("gRPC login called: email={}", request.getEmail());
        try {
            TokenPair tokenPair = userService.login(
                request.getEmail(),
                request.getPassword()
            );
            TokenResponse response = TokenResponse.newBuilder()
                .setAccessToken(tokenPair.getAccessToken())
                .setRefreshToken(tokenPair.getRefreshToken())
                .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            logger.info("gRPC login successful for email={}", request.getEmail());
        } catch (InvalidInputException e) {
            logger.warn("gRPC login invalid input: {}", e.getMessage());
            responseObserver.onError(Status.INVALID_ARGUMENT
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (UserNotFoundException e) {
            logger.warn("gRPC login user not found: {}", e.getMessage());
            responseObserver.onError(Status.NOT_FOUND
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (InvalidCredentialsException e) {
            logger.warn("gRPC login invalid credentials: {}", e.getMessage());
            responseObserver.onError(Status.UNAUTHENTICATED
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (Exception e) {
            logger.error("gRPC login internal error: {}", e.getMessage(), e);
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }

    @Override
    public void refresh(RefreshRequest request, StreamObserver<TokenResponse> responseObserver) {
        logger.info("gRPC refresh called");
        try {
            TokenPair tokenPair = userService.refresh(
                request.getRefreshToken()
            );
            TokenResponse response = TokenResponse.newBuilder()
                .setAccessToken(tokenPair.getAccessToken())
                .setRefreshToken(tokenPair.getRefreshToken())
                .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            logger.info("gRPC refresh successful");
        } catch (InvalidInputException e) {
            logger.warn("gRPC refresh invalid input: {}", e.getMessage());
            responseObserver.onError(Status.INVALID_ARGUMENT
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (UserNotFoundException e) {
            logger.warn("gRPC refresh user not found: {}", e.getMessage());
            responseObserver.onError(Status.NOT_FOUND
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (InvalidCredentialsException e) {
            logger.warn("gRPC refresh invalid credentials: {}", e.getMessage());
            responseObserver.onError(Status.UNAUTHENTICATED
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (Exception e) {
            logger.error("gRPC refresh internal error: {}", e.getMessage(), e);
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }
    
    @Override
    public void logout(LogoutRequest request, StreamObserver<LogoutResponse> responseObserver) {
        logger.info("gRPC logout called");
        try {
            String message = userService.logout(
                request.getRefreshToken()
            );
            LogoutResponse response = LogoutResponse.newBuilder()
                .setMessage(message)
                .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            logger.info("gRPC logout successful");
        } catch (InvalidInputException e) {
            logger.warn("gRPC logout invalid input: {}", e.getMessage());
            responseObserver.onError(Status.INVALID_ARGUMENT
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (InvalidCredentialsException e) {
            logger.warn("gRPC logout invalid credentials: {}", e.getMessage());
            responseObserver.onError(Status.UNAUTHENTICATED
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (Exception e) {
            logger.error("gRPC logout internal error: {}", e.getMessage(), e);
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }

    @Override
    public void getUsersData(GetUsersDataRequest request, StreamObserver<GetUsersDataResponse> responseObserver) {
        logger.info("gRPC getUsersData called for userIds={}", request.getUserIdList());
        try {
            List<String> userIds = request.getUserIdList();
            Map<String, List<String>> userData = userService.getUsersData(userIds);
            List<String> usernames = userData.get("usernames");
            List<String> foundUserIds =  userData.get("userIds");
            GetUsersDataResponse.Builder responseBuilder = GetUsersDataResponse.newBuilder();
            if (usernames != null) {
                responseBuilder.addAllUsername(usernames);
            }
            if (foundUserIds != null) {
                responseBuilder.addAllUserId(foundUserIds);
            }
            responseObserver.onNext(responseBuilder.build());
            responseObserver.onCompleted();
            logger.info("gRPC getUsersData successful for userIds={}", userIds);
        } catch (Exception e) {
            logger.error("gRPC getUsersData internal error: {}", e.getMessage(), e);
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }
}
