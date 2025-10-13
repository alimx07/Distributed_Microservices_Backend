package com.mini_x.follow_service.entity;

import java.time.LocalDateTime;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.IdClass;
import jakarta.persistence.PrePersist;
import jakarta.persistence.Table;

@Entity
@Table(name = "follow")
@IdClass(FollowID.class)
public class Follow {

    @Id
    private Long followerID;

    @Id
    private Long followeeID;


    @Column(name = "created_at" , nullable=false , updatable=false)
    private LocalDateTime createdAt;


    public Follow(long followerID , long followeeID) {
        this.followerID = followerID;
        this.followeeID = followeeID;
    }

    @PrePersist
    public void onCreate() {
        this.createdAt = LocalDateTime.now();
    }

    public Long getFollowerId() {
        return followerID;
    }
    
    public void setFollowerID(Long followerId) {
        this.followerID = followerId;
    }
    
    public Long getFolloweeID() {
        return followeeID;
    }
    
    public void setFolloweeID(Long followingId) {
        this.followeeID = followingId;
    }
    
    public LocalDateTime getCreatedAt() {
        return createdAt;
    }
    
    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }
}
