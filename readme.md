## Project Overview
This repository is a practical microservices backend.


Services:
1 - API GATEWAY : Handles User Auth , RateLimiter (using Redis) , Routing and serivces loadbalancing.
1 - User :  Handles For User data 
2 - Feed :  Generate UserFeed Using Hybird model of fanout_on_write and fanout_on_read and caches
3 - Post :  Hanles Posts Data (posts , comments and likes) using Postgres with replica
4 - Follow : Handles all follow relations
5 - Chat : Responsable for Chatting in the app using websockets

Core goals:
- Build a small but realistic system with multiple services, different techs and pragmatic trade-offs.
- Push it to Kubernetes later, and observe it end-to-end with Prometheus/Grafana.
- Testing the system using K6, measure it, and fix real bottlenecks.

Notes:
 - All services should talk with gRPC and provide Rest Apis for client
 - API_Gateway responsable for converting Https for Grpc and single auth to system