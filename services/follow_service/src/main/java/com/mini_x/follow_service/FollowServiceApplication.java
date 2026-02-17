package com.mini_x.follow_service;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class FollowServiceApplication {

	public static void main(String[] args) {
		SpringApplication.run(FollowServiceApplication.class, args);

		// Application is now running!
		// HTTP server listening on port 8080 (FOR HEALTH CHECK ONLY)
	}

}
