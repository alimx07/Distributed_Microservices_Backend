package com.mini_x.follow_service.service;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.mini_x.follow_service.entity.Follow;
import com.mini_x.follow_service.entity.FollowID;
import com.mini_x.follow_service.exception.FollowAlreadyExistsException;
import com.mini_x.follow_service.exception.FollowNotFoundException;
import com.mini_x.follow_service.exception.InvalidInputException;
import com.mini_x.follow_service.repo.Read.ReadRepo;
import com.mini_x.follow_service.repo.Write.WriteRepo;



/** 
 * Architecture Flow:
 * gRPC Service → Follow Service → Repository → Database
*/
@Service
public class FollowServiceImp implements FollowService {
    
    private final WriteRepo writeRepo;
    private final ReadRepo readRepo;

    private static final long THRESHOLD = 10000;


    // Components will be ingected by spring boot
    public FollowServiceImp(WriteRepo writeRepo, ReadRepo readRepo){
        this.writeRepo = writeRepo;
        this.readRepo = readRepo;
    }



    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public List<String> getFollowers(String userId) {
        List<String> followers = this.readRepo.findFolloweesByFollowerId(userId);
        return followers;
    }
    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public List<String> getFollowing(String userId) {
        List<String> following = this.readRepo.findFollowersByFolloweeId(userId);
        return following;
    }


    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public boolean isCeleb(String userId) {
        long celeb = this.readRepo.countFollowersByFolloweeId(userId);
        return (celeb > THRESHOLD);
    }


    @Override
    @Transactional("primaryTransactionManager")
    public void createFollow(String followerId, String followeeId) {
        if (followerId == null) {
            throw new InvalidInputException("FollowerID can not be Null");
        }
        if (followeeId == null) {
            throw new InvalidInputException("FolloweeID can not be Null");
        }
        if (followerId.equals(followeeId)) {
            throw new InvalidInputException("User can not follow himself");
        }
        if (isFollowing(followerId , followeeId)) {
            throw new FollowAlreadyExistsException(followerId , followeeId);
        }

        Follow follow = new Follow(followerId , followeeId);
        this.writeRepo.save(follow);
    }

    @Override
    @Transactional("primaryTransactionManager")
    public void deleteFollow(String followerId , String followeeId) {
        if (followerId == null) {
            throw new InvalidInputException("FollowerID can not be Null");
        }
        if (followeeId == null) {
            throw new InvalidInputException("FolloweeID can not be Null");
        }
        if (followerId.equals(followeeId)) {
            throw new InvalidInputException("User can not unfollow himself");
        }
        if (!isFollowing(followerId , followeeId)) {
            throw new FollowNotFoundException(followerId , followeeId);
        }
        Follow follow = new Follow(followerId , followeeId);
        this.writeRepo.delete(follow);
    }

    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    private boolean isFollowing(String followerID , String followeeID) {
        
        FollowID id = new FollowID(followerID , followeeID);
        return this.readRepo.existsById(id);
    }
}