package com.mini_x.user_service.config;

import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.data.redis.connection.RedisClusterConfiguration;
import org.springframework.data.redis.connection.RedisNode;
import org.springframework.data.redis.connection.RedisStandaloneConfiguration;
import org.springframework.data.redis.connection.lettuce.LettuceConnectionFactory;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.serializer.GenericJackson2JsonRedisSerializer;
import org.springframework.data.redis.serializer.StringRedisSerializer;

@Configuration
public class RedisConfig {

    @Value("${spring.data.redis.host}")
    private String standaloneHost;

    @Value("${spring.data.redis.port}")
    private int standalonePort;

    @Value("${spring.data.redis.cluster.nodes}")
    private String clusterNodes;

    // Standalone Redis for User Cache
    @Bean(name = "standaloneRedisConnectionFactory")
    @Primary
    public LettuceConnectionFactory standaloneRedisConnectionFactory() {
        RedisStandaloneConfiguration config = new RedisStandaloneConfiguration(standaloneHost, standalonePort);
        return new LettuceConnectionFactory(config);
    }

    @Bean(name = "userCacheRedisTemplate")
    @Primary
    public RedisTemplate<String, Object> userCacheRedisTemplate(
            @Qualifier("standaloneRedisConnectionFactory") LettuceConnectionFactory connectionFactory) {
        RedisTemplate<String, Object> template = new RedisTemplate<>();
        template.setConnectionFactory(connectionFactory);
        template.setKeySerializer(new StringRedisSerializer());
        template.setValueSerializer(new GenericJackson2JsonRedisSerializer());
        template.setHashKeySerializer(new StringRedisSerializer());
        template.setHashValueSerializer(new GenericJackson2JsonRedisSerializer());
        return template;
    }

    // Redis Cluster for Token Management
    @Bean(name = "clusterRedisConnectionFactory")
    public LettuceConnectionFactory clusterRedisConnectionFactory() {
        List<RedisNode> nodes = Arrays.stream(clusterNodes.split(","))
            .map(String::trim)
            .map(node -> {
                String[] parts = node.split(":");
                return new RedisNode(parts[0], Integer.parseInt(parts[1]));
            })
            .collect(Collectors.toList());
        
        RedisClusterConfiguration clusterConfiguration = new RedisClusterConfiguration();
        clusterConfiguration.setClusterNodes(nodes);
        
        return new LettuceConnectionFactory(clusterConfiguration);
    }

    @Bean(name = "tokenCacheRedisTemplate")
    public RedisTemplate<String, Object> tokenCacheRedisTemplate(
            @Qualifier("clusterRedisConnectionFactory") LettuceConnectionFactory connectionFactory) {
        RedisTemplate<String, Object> template = new RedisTemplate<>();
        template.setConnectionFactory(connectionFactory);
        template.setKeySerializer(new StringRedisSerializer());
        template.setValueSerializer(new GenericJackson2JsonRedisSerializer());
        template.setHashKeySerializer(new StringRedisSerializer());
        template.setHashValueSerializer(new GenericJackson2JsonRedisSerializer());
        return template;
    }
}