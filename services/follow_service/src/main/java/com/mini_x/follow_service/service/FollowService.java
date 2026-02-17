package com.mini_x.follow_service.service;


import java.util.List;



public interface FollowService {

    List<String> getFollowers(String userId);

    List<String> getFollowing(String userId);

    boolean isCeleb(String userId);

    void createFollow(String followerId , String FolloweeId);

    void deleteFollow(String followerId , String FolloweeId);
    
}