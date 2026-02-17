package com.mini_x.follow_service.repo.Read;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import com.mini_x.follow_service.entity.Follow;
import com.mini_x.follow_service.entity.FollowID;


@Repository
@Transactional(value = "secondaryTransactionManager", readOnly = true)
public interface ReadRepo extends  JpaRepository<Follow, FollowID>{


    @Query("SELECT f.followerID FROM Follow f WHERE f.followeeID = :userId")
    List<String> findFollowersByFolloweeId(@Param("userId") String userId);
    

    @Query("SELECT f.followeeID FROM Follow f WHERE f.followerID = :userId")
    List<String> findFolloweesByFollowerId(@Param("userId") String userId);
    
    @Query("SELECT COUNT(f) FROM Follow f WHERE f.followeeID = :userId")
    long countFollowersByFolloweeId(@Param("userId") String userId);
}
