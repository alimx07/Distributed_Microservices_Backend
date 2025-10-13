package com.mini_x.follow_service.entity;

import java.io.Serializable;
import java.util.Objects;
public class FollowID implements Serializable {

    private long followerID;
    private long followeeID; 

    public FollowID() {}
    
    public FollowID(long followerID , long followeeID) {
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

    public long getFollowerID(){
        return followerID;
    }

    public void setFollowerID(long x) {
        this.followerID = x;
    }

    public long getFolloweeID(){
        return followeeID;
    }

    public void setFolloweeID(long x) {
        this.followeeID = x;
    }
    
}
