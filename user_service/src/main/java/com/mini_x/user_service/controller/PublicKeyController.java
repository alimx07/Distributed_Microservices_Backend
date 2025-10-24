package com.mini_x.user_service.controller;

import java.util.HashMap;
import java.util.Map;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class PublicKeyController {

    private final String publicKeyPem;

    public PublicKeyController(String publicKeyPem) {
        this.publicKeyPem = publicKeyPem;
    }
    @GetMapping("/public-key")
    public Map<String, String> getPublicKey() {
        Map<String, String> resp = new HashMap<>();
        resp.put("publicKey", publicKeyPem);
        return resp;
    }
}
