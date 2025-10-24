package com.mini_x.user_service.repo.Read;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import com.mini_x.user_service.entity.User;

@Repository
@Transactional(value = "secondaryTransactionManager", readOnly = true)
public interface ReadRepo extends JpaRepository<User, String> {

    @Query("SELECT u FROM User u WHERE u.email = :email")
    Optional<User> findByEmail(@Param("email") String email);

    @Query("SELECT u FROM User u WHERE u.userid IN :userIds")
    List<User> findUsersByIds(@Param("userIds") List<String> userIds);
}
