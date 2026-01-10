# Overview

There’s a common saying in the industry that the first rule of microservices is: *don’t do microservices*. The idea is that if you can avoid the complexity, you probably should. In this repository, I do them anyway — but with a focus on understanding the real cost and learning.

This is not just about writing code. It includes the system design around the services, the infrastructure including data stores, message queues, caching layers, and other supporting pieces that make distributed systems actually work.

The journey is incremental and covers several stages:

1. **Services Desing** — starting with Design & implement the services.
2. **Containerization** — packaging services and dependencies with Docker Compose, managing networks, secrets, and inter-service communication.
3. **Kubernetes** — moving toward production-style deployment, with service discovery, networking, scaling, and security in mind.
4. **Observability on Kubernetes** — adding logging, metrics, and tracing to understand behavior and failures across services.


## Services

The project will include multiple services written in **Go** and **Spring Boot**, including:

- `api_gateway`
- `post_service`
- `follow_service`
- `feed_service`
- `user_service`

---------------------------------------------

Work is in progress. Iterations will be pushed as the system evolves.


----------------------------------------------


## Services Design & Implementation
