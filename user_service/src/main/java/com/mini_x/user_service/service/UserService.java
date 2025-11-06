package com.mini_x.user_service.service;

import java.util.List;
import java.util.Map;

import com.mini_x.user_service.dto.TokenPair;

public interface UserService {

    void register(String username, String email, String password);

    TokenPair login(String email, String password);
    
    TokenPair refresh(String refreshToken);
    
    String logout(String refreshToken);

    Map<String, List<String>> getUsersData(List<String> userIds);
}
