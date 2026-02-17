package com.mini_x.user_service.config;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.security.KeyFactory;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.spec.PKCS8EncodedKeySpec;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.io.Resource;

@Configuration
public class JwtConfig {

    @Value("${jwt.private-key-path:classpath:keys/private.pem}")
    private Resource privateKeyResource;

    @Value("${jwt.public-key-path:classpath:keys/public.pem}")
    private Resource publicKeyResource;

    @Bean
    public PrivateKey jwtPrivateKey() throws Exception {
        String privateKeyPEM = new String(Files.readAllBytes(Paths.get(privateKeyResource.getURI())))
            .replace("-----BEGIN PRIVATE KEY-----", "")
            .replace("-----END PRIVATE KEY-----", "")
            .replaceAll("\\s+", "");

        byte[] decoded = Base64.getDecoder().decode(privateKeyPEM);
        
        PKCS8EncodedKeySpec spec = new PKCS8EncodedKeySpec(decoded);
        KeyFactory kf = KeyFactory.getInstance("RSA");
        return kf.generatePrivate(spec);
    }

    @Bean
    public PublicKey jwtPublicKey() throws Exception {
        String publicKeyPEM = new String(Files.readAllBytes(Paths.get(publicKeyResource.getURI())))
            .replace("-----BEGIN PUBLIC KEY-----", "")
            .replace("-----END PUBLIC KEY-----", "")
            .replaceAll("\\s+", "");

        byte[] decoded = Base64.getDecoder().decode(publicKeyPEM);
        
        X509EncodedKeySpec spec = new X509EncodedKeySpec(decoded);
        KeyFactory kf = KeyFactory.getInstance("RSA");
        return kf.generatePublic(spec);
    }
    
    @Bean
    public String publicKeyPem() throws IOException {
        return new String(Files.readAllBytes(Paths.get(publicKeyResource.getURI())));
    }
}
