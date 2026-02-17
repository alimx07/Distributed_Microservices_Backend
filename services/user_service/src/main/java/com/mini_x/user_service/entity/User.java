package com.mini_x.user_service.entity;

import java.time.LocalDateTime;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.PrePersist;
import jakarta.persistence.Table;
import com.github.f4b6a3.ulid.UlidCreator;

@Entity
@Table(name = "users")
public class User {

    @Id
    @Column(name = "id" , nullable=false , unique=true , length=26) // ULID
    private String userid;

    @Column(name = "username", nullable = false, unique = true, length = 50)
    private String username;

    @Column(name = "email", nullable = false, unique = true, length = 100)
    private String email;

    @Column(name = "password", nullable = false, length = 100)
    private String password;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    public User() {
        this.userid = UlidCreator.getUlid().toString();
    }

    public User(String username, String email, String password) {
        this.userid = UlidCreator.getUlid().toString();
        this.username = username;
        this.email = email;
        this.password = password;
    }

    @PrePersist
    public void onCreate() {
        if (this.userid == null) {
            this.userid = UlidCreator.getUlid().toString();
        }
        this.createdAt = LocalDateTime.now();
    }

    public String getUserid() {
        return userid;
    }

    public void setUserid(String userid) {
        this.userid = userid;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }
}
