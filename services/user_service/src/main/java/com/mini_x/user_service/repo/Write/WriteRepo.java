package com.mini_x.user_service.repo.Write;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import com.mini_x.user_service.entity.User;

@Repository
@Transactional("primaryTransactionManager")
public interface WriteRepo extends JpaRepository<User, String> {
    
    /*
     * ONLY NEED TO WRITE
     * HIBERNATE METHODS ARE ENOUGH NOW
     */
}
