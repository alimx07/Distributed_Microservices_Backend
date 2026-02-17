package com.mini_x.follow_service.grpc;

import java.util.List;

import org.springframework.stereotype.Component;

import com.mini_x.follow_service.exception.FollowAlreadyExistsException;
import com.mini_x.follow_service.exception.FollowNotFoundException;
import com.mini_x.follow_service.exception.InvalidInputException;
import com.mini_x.follow_service.service.FollowService;

import io.grpc.stub.StreamObserver;
import net.devh.boot.grpc.server.service.GrpcService;

@Component
@GrpcService
public class FollowGrpcService extends FollowServiceGrpc.FollowServiceImplBase {

    private final FollowService followService;

    public FollowGrpcService(FollowService followService) {
        this.followService = followService;
    }


    @Override
    public void getFollowers(GetFollowersReq request, StreamObserver<GetFollowersRes> responseObserver) {
        try {
            String userId = request.getUserId();
            List<String> followers = followService.getFollowers(userId);
            
            GetFollowersRes response = GetFollowersRes.newBuilder()
                    .addAllFollowerID(followers)
                    .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                    .withDescription("Error getting followers: " + e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }


    @Override
    public void getCeleb(GetFollowersReq request, StreamObserver<GetFollowersRes> responseObserver) {
        try {
            String userId = request.getUserId();
            // Get following list (people this user follows)
            List<String> following = followService.getFollowing(userId);
            
            GetFollowersRes response = GetFollowersRes.newBuilder()
                    .addAllFollowerID(following)
                    .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                    .withDescription("Error getting celeb followers: " + e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }


    @Override
    public void isCeleb(IsCelebReq request, StreamObserver<IsCelebRes> responseObserver) {
        try {
            String userId = request.getUserId();
            boolean isCelebrity = followService.isCeleb(userId);
            
            IsCelebRes response = IsCelebRes.newBuilder()
                    .setIsCeleb(isCelebrity)
                    .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                    .withDescription("Error checking celebrity status: " + e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }


    @Override
    public void createFollow(CreateFollowReq request, StreamObserver<FollowResponse> responseObserver) {
        try {
            String followerId = request.getUserId();
            String followeeId = request.getFolloweeId();
            
            followService.createFollow(followerId, followeeId);
            
            FollowResponse response = FollowResponse.newBuilder()
                    .setMessage("Successfully followed user")
                    .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (InvalidInputException e) {
            responseObserver.onError(io.grpc.Status.INVALID_ARGUMENT
                    .withDescription(e.getMessage())
                    .asRuntimeException());
        } catch (FollowAlreadyExistsException e) {
            responseObserver.onError(io.grpc.Status.ALREADY_EXISTS
                    .withDescription(e.getMessage())
                    .asRuntimeException());
        } catch (Exception e) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                    .withDescription("Error creating follow: " + e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }


    @Override
    public void deleteFollow(DeleteFollowReq request, StreamObserver<FollowResponse> responseObserver) {
        try {
            String followerId = request.getUserId();
            String followeeId = request.getFolloweeId();
            
            followService.deleteFollow(followerId, followeeId);
            
            FollowResponse response = FollowResponse.newBuilder()
                    .setMessage("Successfully unfollowed user")
                    .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (InvalidInputException e) {
            responseObserver.onError(io.grpc.Status.INVALID_ARGUMENT
                    .withDescription(e.getMessage())
                    .asRuntimeException());
        } catch (FollowNotFoundException e) {
            responseObserver.onError(io.grpc.Status.NOT_FOUND
                    .withDescription(e.getMessage())
                    .asRuntimeException());
        } catch (Exception e) {
            responseObserver.onError(io.grpc.Status.INTERNAL
                    .withDescription("Error deleting follow: " + e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }
}
