package com.mini_x.follow_service.service;


import java.util.List;



public interface FollowService {

    List<Long> getFollowers(Long userId);

    List<Long> getFollowing(Long userId);

    boolean isCeleb(Long userId);

    void createFollow(Long followerId , Long FolloweeId);

    void deleteFollow(Long followerId , Long FolloweeId);
    
}