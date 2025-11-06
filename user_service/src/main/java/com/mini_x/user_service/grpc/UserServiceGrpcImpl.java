package com.mini_x.user_service.grpc;

import java.util.List;
import java.util.Map;

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

    private final UserService userService;

    public UserServiceGrpcImpl(UserService userService) {
        this.userService = userService;
    }

    @Override
    public void register(RegisterRequest request, StreamObserver<RegisterResponse> responseObserver) {
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

        } catch (InvalidInputException e) {
            responseObserver.onError(Status.INVALID_ARGUMENT
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (UserAlreadyExistsException e) {
            responseObserver.onError(Status.ALREADY_EXISTS
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }

    @Override
    public void login(LoginRequest request, StreamObserver<TokenResponse> responseObserver) {
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

        } catch (InvalidInputException e) {
            responseObserver.onError(Status.INVALID_ARGUMENT
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (UserNotFoundException e) {
            responseObserver.onError(Status.NOT_FOUND
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (InvalidCredentialsException e) {
            responseObserver.onError(Status.UNAUTHENTICATED
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }

    @Override
    public void refresh(RefreshRequest request, StreamObserver<TokenResponse> responseObserver) {
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
            
        } catch (InvalidInputException e) {
            responseObserver.onError(Status.INVALID_ARGUMENT
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (UserNotFoundException e) {
            responseObserver.onError(Status.NOT_FOUND
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (InvalidCredentialsException e) {
            responseObserver.onError(Status.UNAUTHENTICATED
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }
    
    @Override
    public void logout(LogoutRequest request, StreamObserver<LogoutResponse> responseObserver) {
        try {
            String message = userService.logout(
                request.getRefreshToken()
            );

            LogoutResponse response = LogoutResponse.newBuilder()
                .setMessage(message)
                .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
            
        } catch (InvalidInputException e) {
            responseObserver.onError(Status.INVALID_ARGUMENT
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (InvalidCredentialsException e) {
            responseObserver.onError(Status.UNAUTHENTICATED
                .withDescription(e.getMessage())
                .asRuntimeException());
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }

    @Override
    public void getUsersData(GetUsersDataRequest request, StreamObserver<GetUsersDataResponse> responseObserver) {
        try {
            List<String> userIds = request.getUseridList();
            Map<String, List<String>> userData = userService.getUsersData(userIds);

            
            List<String> usernames = userData.get("usernames");
            List<String> foundUserIds =  userData.get("userIds");

            GetUsersDataResponse.Builder responseBuilder = GetUsersDataResponse.newBuilder();
            
            if (usernames != null) {
                responseBuilder.addAllUsername(usernames);
            }
            if (foundUserIds != null) {
                responseBuilder.addAllUserID(foundUserIds);
            }

            responseObserver.onNext(responseBuilder.build());
            responseObserver.onCompleted();

        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                .withDescription("Internal server error: " + e.getMessage())
                .asRuntimeException());
        }
    }
}
