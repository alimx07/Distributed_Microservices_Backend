package com.mini_x.follow_service.repo.Write;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import com.mini_x.follow_service.entity.Follow;
import com.mini_x.follow_service.entity.FollowID;

@Repository
@Transactional("primaryTransactionManager")
public interface WriteRepo extends JpaRepository<Follow, FollowID>{


    /*
     * ONLY NEED TO WRITE AND DELETE
     * HIBERNATE METHODS ARE ENOUGH NOW
     */
}
