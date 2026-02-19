package com.mini_x.user_service.config;

// import io.etcd.jetcd.Client;
// import org.springframework.beans.factory.annotation.Value;
// import org.springframework.context.annotation.Bean;
// import org.springframework.context.annotation.Configuration;

// @Configuration
// public class EtcdConfig {

//     @Value("${etcd.endpoints}")
//     private String etcdEndpoints;

//     @Bean
//     public Client etcdClient() {
//         String[] endpoints = etcdEndpoints.split(",");
//         return Client.builder()
//                 .endpoints(endpoints)
//                 .build();
//     }
// }
