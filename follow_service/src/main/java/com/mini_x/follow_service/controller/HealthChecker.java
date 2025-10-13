package com.mini_x.follow_service.controller;

import java.util.HashMap;
import java.util.Map;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class HealthChecker {

    @GetMapping("/health")
    public Map<String, String> health() {
        Map<String , String> resp = new HashMap<>();

        resp.put("status" , "up");
        resp.put("service" , "follow_service");

        return resp;
    }
}
