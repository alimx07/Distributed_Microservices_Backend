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
    @Column(name = "follower_id", nullable = false, length = 26)
    private String followerID;

    @Id
    @Column(name = "followee_id", nullable = false, length = 26)
    private String followeeID;


    @Column(name = "created_at" , nullable=false , updatable=false)
    private LocalDateTime createdAt;

   
    public Follow() {}

    public Follow(String followerID , String followeeID) {
        this.followerID = followerID;
        this.followeeID = followeeID;
    }

    @PrePersist
    public void onCreate() {
        this.createdAt = LocalDateTime.now();
    }

    public String getFollowerID() {
        return followerID;
    }
    
    public void setFollowerID(String followerID) {
        this.followerID = followerID;
    }
    
    public String getFolloweeID() {
        return followeeID;
    }
    
    public void setFolloweeID(String followeeID) {
        this.followeeID = followeeID;
    }
    
    public LocalDateTime getCreatedAt() {
        return createdAt;
    }
    
    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }
}
