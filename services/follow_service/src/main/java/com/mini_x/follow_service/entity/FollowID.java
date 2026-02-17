package com.mini_x.follow_service.entity;

import java.io.Serializable;
import java.util.Objects;
public class FollowID implements Serializable {

    private String followerID;
    private String followeeID; 

    public FollowID() {}
    
    public FollowID(String followerID , String followeeID) {
        this.followerID = followerID;
        this.followeeID = followeeID;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        FollowID that = (FollowID) o;
        return Objects.equals(followerID, that.followerID) && Objects.equals(followeeID, that.followeeID);
    }

    @Override
    public int hashCode() {
        return Objects.hash(followerID, followeeID);
    }

    public String getFollowerID(){
        return followerID;
    }

    public void setFollowerID(String x) {
        this.followerID = x;
    }

    public String getFolloweeID(){
        return followeeID;
    }

    public void setFolloweeID(String x) {
        this.followeeID = x;
    }
    
}
