## Project Overview
This repository is a practical microservices backend.


Services:
- API Gateway
  - User authentication
  - Rate limiting (Redis)
  - Routing and service load balancing
  - REST/HTTP to gRPC translation

- User Service
  - User data management
  - PostgreSQL with a read replica
- Feed Service
  - User feed generation using a hybrid of fanout-on-write and fanout-on-read with caches

- Post Service
  - Posts, comments, likes
  - PostgreSQL with a read replica
  - Outbox Pattern by Using Debezium and Kafka

- Follow Service
  - Follow relationships

- Chat Service
  - WebSocket-based chat

Core goals:
- Build a small but realistic system with multiple services, different techs and pragmatic trade-offs.
- Push it to Kubernetes later, and observe it end-to-end with Prometheus/Grafana.
- Testing the system using K6, measure it, and fix real bottlenecks.

Notes:
 - All services should talk with gRPC and provide Rest Apis for client
 - API_Gateway responsable for converting Https for Grpc and single auth to system