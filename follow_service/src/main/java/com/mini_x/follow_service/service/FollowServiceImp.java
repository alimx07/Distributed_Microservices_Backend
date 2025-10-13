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

    private static final long Threshold = 10000;


    // Components will be ingected by spring boot
    public FollowServiceImp(WriteRepo writeRepo, ReadRepo readRepo){
        this.writeRepo = writeRepo;
        this.readRepo = readRepo;
    }



    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public List<Long> getFollowers(Long userId) {
        List<Long> followers = this.readRepo.findFolloweesByFollowerId(userId);
        return followers;
    }
    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public List<Long> getFollowing(Long userId) {
        List<Long> following = this.readRepo.findFollowersByFolloweeId(userId);
        return following;
    }

    @Override
    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    public boolean isCeleb(Long userId) {
        long celeb = this.readRepo.countFollowersByFolloweeId(userId);
        return (celeb > Threshold);
    }


    @Override
    @Transactional("primaryTransactionManager")
    public void createFollow(Long followerId, Long followeeId) {
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
    public void deleteFollow(Long followerId , Long followeeId) {
        if (followerId == null) {
            throw new InvalidInputException("FollowerID can not be Null");
        }
        if (followeeId == null) {
            throw new InvalidInputException("FolloweeID can not be Null");
        }
        if (!followerId.equals(followeeId)) {
            throw new InvalidInputException("User can not follow himself");
        }
        if (!isFollowing(followerId , followeeId)) {
            throw new FollowNotFoundException(followerId , followeeId);
        }
        Follow follow = new Follow(followerId , followeeId);
        this.writeRepo.delete(follow);
    }

    @Transactional(value = "secondaryTransactionManager", readOnly = true)
    private boolean isFollowing(Long followerID , Long followeeID) {
        
        FollowID id = new FollowID(followerID , followeeID);
        return this.writeRepo.existsById(id);
    }
}