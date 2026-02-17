package com.mini_x.user_service.controller;

import java.util.HashMap;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class PublicKeyController {

    private static final Logger logger = LoggerFactory.getLogger(PublicKeyController.class);
    
    private final String publicKeyPem;

    public PublicKeyController(String publicKeyPem) {
        this.publicKeyPem = publicKeyPem;
        logger.info("PublicKeyController initialized");
        if (publicKeyPem == null || publicKeyPem.isEmpty()) {
            logger.error("Public key PEM is null or empty during initialization");
        } else {
            logger.debug("Public key loaded successfully, length: {}", publicKeyPem.length());
        }
    }
    
    @GetMapping("/public-key")
    public Map<String, String> getPublicKey() {
        logger.info("Received request for public key");
        
        try {
            Map<String, String> resp = new HashMap<>();
            
            if (publicKeyPem == null || publicKeyPem.isEmpty()) {
                logger.error("Public key is null or empty when serving request");
                resp.put("error", "Public key not available");
                return resp;
            }
            
            resp.put("publicKey", publicKeyPem);
            logger.info("Successfully returned public key, key length: {}", publicKeyPem.length());
            
            return resp;
        } catch (Exception e) {
            logger.error("Error occurred while serving public key request", e);
            Map<String, String> errorResp = new HashMap<>();
            errorResp.put("error", "Internal server error: " + e.getMessage());
            return errorResp;
        }
    }
}
