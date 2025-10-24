package com.mini_x.user_service;


import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class UserServiceApplication {

	public static void main(String[] args) {
		SpringApplication.run(UserServiceApplication.class, args);

		// Application is now running!
		// HTTP server listening on port 8080 (FOR HEALTH CHECK && Public Key)
		// gRPC server will be configured via application.properties

		// JAVA IS WILD (TOOOOOO MUCH)
	}

}
