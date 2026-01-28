# Overview

There’s a common saying in the industry that the first rule of microservices is: **don’t do microservices**. The idea is that if you can avoid the complexity, you probably should. In this repository, I do them anyway (even if it is just a mini example) — but with a focus on understanding the real cost and learning.

This is not just about writing code. It includes the system design around the services, the infrastructure , deploying and other supporting pieces that make distributed systems actually work.

The journey is incremental and covers several stages:

1. **Services Design** — starting with Design & implement the services.
2. **Deployment on AWS ECS** — Using Terrafrom and Terragrunt (DRY PATTERN & DIFFERENT ENVs)
3. **Migration to AWS EKS** 
4. **Observability on EKS** — logging/metrics/tracing


## Services

The project be a social media platfrom will include multiple services from scratch written in **Go** and **Spring Boot**, including:

- `api_gateway`
- `post_service`
- `follow_service`
- `feed_service`
- `user_service`

---------------------------------------------

Work is in progress. Iterations will be pushed as the system evolves.


## Table of Contents

- [API Gateway](#api-gateway)
  - [Rate Limiting](#rate-limiting)
  - [Auth](#auth)
  - [HTTP to gRPC](#http-to-grpc)
  - [Load Balancer](#load-balancer)
- [Post Service](#post-service)
  - [CDC + Kafka + Outbox Pattern](#cdc--kafka--outbox-pattern)
- [Feed Service](#feed-service)
  - [Fan-out on Write](#1-fan-out-on-write)
  - [Fan-out on Read](#2-fan-out-on-read)
- [User Service](#user-service)
- [Follow Service](#follow-service)

----------------------------------------------

![Full Service Design](images/fullDesign.png)

## API Gateway


Different API Gateways found like Kong , Traefik , HAProxy, but the motivation here to implement a dynamic one from scratch , yes not production ready but it will do a basic and important operations needed in my design.

The Gateway basic needs are:
- Dynamic configs
- rate limiting (using redis & lua scripts)
- Auth
- convert HTTP to GRPC (Route to specific service)
- load balancing



Let`s start with the path of any request:


## Rate Limiting

We initialize the **rateLimiter** with our configuration.  
Whenever a new request arrives, we check whether it should be rate-limited:

- If **no**, proceed to the next stage (e.g., Auth)
- If **yes**, check whether the request exceeds the limit

This leads to three questions:

1. What is rate limiting?
2. What do we limit?
3. How do we check the limit?



### What Is Rate Limiting?

> Rate limiting is controlling how many requests can be made by an entity (eg: user) to a resource within a specified period of time.

Without rate limiting, a user could make an excessive number of requests (e.g., `1000 requests/sec`), consuming server resources, degrading service for others. Also, it is used to reduce the impact of Application layer DoS attacks.



### What Do We Limit?

Rate limits can be applied on different dimensions:

- IP address
- User ID
- Endpoint (per user)
- HTTP method
- Combinations of the above

Example rules:

- `1000 requests/min per IP`
- `100 requests/min per User`
- `20 requests/min per UserID + URL`

<!-- > **Note:** Limiting by IP only can be unreliable in our situation because multiple users may share a single IP (e.g., behind NAT), allowing one user to affect others. -->



### How Do We Check the Limit?

There are several algorithms for enforcing rate limits. One common approach is the **Token Bucket** algorithm.

![Token bucket algorithm](images/TokenBucket.png)

#### Token Bucket 

- Define a **bucket size (N)** and a **refill rate**
- Tokens are added at a fixed rate
- Each incoming request consumes one token
- If the bucket has tokens → allow the request
- If empty → reject or delay the request

#### Why Token Bucket ?

Compared to other popular rate limiting algorithms like **Leaky Bucket** and **Sliding Window**, the **Token Bucket** has clear pros for handling user interactions (posts, likes, and comments).

- **Token Bucket**  
  The basic idea of the token bucket allows bursts of user activity up to the number of tokens available in the bucket within a time window, which means users can perform multiple actions quickly without being blocked, naking it in other words *natural for our interactions*.

- **Leaky Bucket**  
  Leaky bucket works by keeping a queue of requests and processing them at a fixed rate. While this regulates the overall rate of requests hitting the server, a sudden burst of traffic can fill the queue with older requests. If these aren’t processed in time, **recent requests may get blocked**, which can affect user experience.

- **Sliding Window**  
  Sliding window tracks requests per window very efficiently and provides exact counting (using timestamps in redis or logs). However, it generally **requires more memory**  compared to token bucket, making it less lightweight for high-traffic social media platforms.


We can implement token bucket logic in the code but this is problematic in distributed environments. As of course we will have multiple gateways , So imagine with me :

#### Example Problem (Multiple Gateways)

Assume multiple API gateways: `A`, `B`, and `C`.

Scenario:

Bucket Size = 10

1. User-1 request hits gateway **A** → Decrement so  `curr_tokens = 9`
2. Next request routes to gateway **B** → Decrement so `curr_tokens = 9` but it should be `8` (incorrect)

This causes inconsistency.

Even if we route all specific user requests so the same gateway. It is still a problem in a situation like what if some gateway crash. now requests go to new gateway which will reset everything assuming it is a new user 
> yes I know it could be just some rate per min, Someone can say it is not a big deal but imagine if it is a rule per hour, day , week. Unfortunately, it is a big deal now



Now if there is a global state (every gateway will call) we can use with a fast access, this would solve our problem. Fortunately, it is found it called **cache**

Redis (popular cache) will provide:

- **Shared global state** across all gateways
- **High performance** suitable for fast checks (of course not as fast as in memory maps)
- **High availability** through clustering
- **Persistence** options for long-term limits


#### Why Lua Scripts?

The Last part of our solution here is to use Lua Scripts. 


But Why do we use Lua scripts in Redis instead of plain Redis commands?

Redis uses a single-threaded command execution engine. All client requests go through a main event loop, which handles I/O, parses commands, and executes them atomically.

When a client sends a request:

- The request is read into a client buffer.
- Redis parses the command using RESP protocol.
- The command is looked up in Redis’ command table.
- It executes the command against the in-memory data structures.
- The response is written back to the client.


Redis commands are executed **atomically**, meaning each single command (like `GET`, `SET`, `INCR`) completes fully before the next command runs.  
**Problem arises** when multiple clients perform **multi-step operations** in separate commands. For example:

  ```text
  Client 1: GET counter
  Client 1: increment value locally
  Client 1: SET counter new_value

  Client 2: GET counter
  Client 2: increment value locally
  Client 2: SET counter new_value
  ```
If both clients read the same value at the same time, they might overwrite each other’s updates, causing lost increments.

This is a classic race condition because the operations are not atomic across multiple commands.


The Solution could be to do the logic in the same request instead of locally. Using that redis can run Lua and every lua script will run atomically (as plain command). Lua scripts allow you to combine multiple steps into one atomic operation on the server:

```lua
local value = redis.call("GET", "counter")
value = value + 1
redis.call("SET", "counter", value)
```


#### Resources
- [Different RateLimiting Algorithms](https://blog.algomaster.io/p/rate-limiting-algorithms-explained-with-code) - AlgoMaster Newspaper
- [X Rate Limits](https://developer.x.com/en/docs/x-api/v1/rate-limits) - X Developer Platform
- [Design a Rate Limiter](https://bytebytego.com/courses/system-design-interview/design-a-rate-limiter) - ByteByteGO

## Auth

After the request passes basic checks, we need to see whether it’s allowed to do the action. To move through this quickly, we need to understand **JWT**.

### What is JWT?

JWT (JSON Web Token) is a signed token containing information about the user that can be used to authenticate them.

> We Store the token in a **cookie**, so the browser automatically sends it with each request. We usually set flags like `httpOnly` and `sameSite` to reduce risks from attacks such as XSS and CSRF.


Access JWTs are usually **short-lived** (e.g., 10 minutes). When they expire, we don’t want to force users to log in again, so we use a **refresh token**. The refresh token can request a new access JWT silently, extending the session without bothering the user.


The idea of using JWT in auth is that instead of storing a session ID in a database and checking it on every request, we can validate a stateless JWT using a public key. No DB call, faster, more scalable.

### But it has a catch

But It does not make sense when we need to **invalidate** users.

Example: 

user logs out or credentials get leaked.  
We can delete the refresh token, but any existing JWT is still valid until it expires.

So if the JWT has 8 minutes left, the attacker keeps access for 8 minutes. That’s not great.

The common fix is having a **revocation list** (a store that keeps track of invalid JWTs). But once you do that, you’re checking a stateful store on each request again. At that point, it’s not very different from session IDs, except JWTs are bigger and carry more bandwidth overhead.

We can store revoked IDs in a global cache again or pushig all invalid UserIDs using Queue to our running gateways and delete it automatically using timer after N mins(Much faster than cache).

#### Where JWTs shine

JWTs are still good for scenarios where access needs a **time window** and doesn’t require logout behavior, like:

- Temporary download links (valid for 3 hours)
- One-time API access for specific tasks


> OAuth providers use tokens that often look like JWTs, but the whole flow and security model is different. If you’re building your own auth system, plain **session IDs with caching** can be more practical and less error-prone. You get revocation, logout, and tracking  JWT’s more size(Avg JWT size 1KB vs Avg Session Size 60B).


<!-- I used Redis to implement the state and using lua scripts for my operations (I can use redis commands here also it will be enough) -->

### Final Auth Logic in my gateway

Final auth flow usually looks like:

1. Check if the URL needs authentication.
2. If yes, get the JWT.
3. Verify the signature.
4. Ensure it isn’t revoked.
5. If all clear, move on (e.g., from HTTP to gRPC). 

> So if JWT is bad, why am I using it? Actually, I didn’t know a lot about it, so I started just for practice. While searching, I found this info and reasons, and I’m lazy to change everything, so I just keep it. But the main idea is JWT for auth + revoke (in most situations) =  bigger session IDs.  

**I may be wrong, so I’ll search more and modify that later.**


## HTTP to gRPC

Now it’s time: our client speaks **HTTP**, but our services speak **gRPC**. How do we translate between them?

The basic idea is to configure endpoints and write a handler for each one that calls the corresponding gRPC method. Sure
```go
func (g *GrpcHanlder) login(w http.ResponseWriter, r *http.Request) {
	
	var reqData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

        // prepare the input message
        grpcRep = pb.LoginReques{}
	resp, err := g.userService.login(grpcRep)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prepare response for client
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
```

that works, But anyone can see this approach is **not scalable**. You end up repeating the same logic for every endpoint. We need a **dynamic approach**.



### Main Idea

We compile our proto files into a **protoset** 
> A protoset is a compiled description of .proto files, stored in a binary format. which contains all the metadata about messages, enums, services, and methods defined in your .proto files. unlike .proto files which are in human-readable text, protoset files are in machine language make it easier to work with tools.

then read our protosetFiles them to know:

- What methods exist in every service.
- The input and output types for each method.

We store this information in our code for later use.

#### Service Map 

The Service Map stores input and output types for every method. This allows the gateway to **create empty Protobuf messages dynamically** when needed.

**Example:**  
> "The method `GetUser` takes a `GetUserRequest` and returns `UserResponse`."

Now the gateway knows what methods(services) exist. But how do we **route suitable requests suitable method**?

No magic here. We can use an extension which is **`google.api.http`** options in our proto files to define HTTP endpoints for gRPC methods.

> **google.api.http** :  is a Protobuf method option (an annotation) that attaches an HttpRule to a gRPC RPC method. Conceptually: “this RPC can also be reached via HTTP/JSON at this verb + URL template”.

**Example:**
```proto
rpc DeletePost(DeletePostRequest) returns (Response) {
    option (google.api.http) = {
        delete: "/api/v1/posts/{PostId}"
    };
}
```

#### Route Map

The Route Map in our code reads `google.api.http` extensions in proto options and dynamically maps **HTTP method + URL → gRPC service method**. This allows the gateway to know exactly which gRPC method to call for each incoming HTTP request.


```
Follow of requests


HTTP Client
   |
   v
[HTTP Gateway Listener]
   |
   |-- lookup Route Map (method + path) --> finds (Service, Method)
   |-- build request message:
   |     - map path vars -> fields
   |     - map query params -> fields
   |     - body JSON -> fields
   |
   |-- lookup Service Map (input/output types) --> get Protobuf descriptor
   |
   |
   |-- Unmarshal requestJSON ---> reqMsg
   |
   v
[gRPC Client] ---> [gRPC Server/Service]
   ^                     |
   |                     v
   |                Protobuf Response
   |
   |-- convert response -> JSON + set HTTP status/headers
   |
   v
HTTP Response -> Client
```


## Load Balancer

Now we actually hit the end but still a small thing **LOAD BALANCING**


In the load balancer part of the API_Gateway, I keep connections with each service instance and routing requests in a **round-robin** manner. This means that the first request goes to Server A, the next request goes to Server B, the next to Server C, and then it repeats back to Server A and so on. 


A key question here is how does the API gateway know which servers are available for a particular service? 

The answer is a **service registry**. The basic idea is simple: 

when a new service instance spins up, it registers itself in the registry. The API gateway can then either pull the list of services or receive updates from the registry when new services are added.  

In my local dev, I used **etcd** (a key-value store), to implement the service registry. This keeps track of all the running service instances and push them to api_gateway so the load balancer can route requests correctly. when it is time to something like AWS, the setup for the service registry might be different with same idea, overall this is a story for another day.




## Post Service

So now it's time to talk about Services.I will start with **POST SERVICE**. Basically, whenever we want to do any operation related to posts (create post, add post, likes, comments, etc.), all of that will be handled inside the post service. The main idea is pretty straightforward: the post service owns everything related to posts.

### Core Components of the Post Service

The post service talks to a Postgres database. I manually configured a replica for high availability by creating a physical replication slot and replication user (yeah, I know there are tools do that for Postgres, but the whole point here is to do things from scratch).

```postgres 
CREATE ROLE physical_rep WITH REPLICATION LOGIN PASSWORD 'YOUR_PASSWORD';

SELECT pg_create_physical_replication_slot('secondary_slot');
```

Also a configuration is modified in postgres.conf at the first intiallization of it by a bash script

```sh
#!/usr/bin/env bash
set -e


echo "Configuring primary for streaming replication..."

cat >> "$PGDATA/postgresql.conf" <<'EOF'
# replication settings
listen_addresses = '*'
wal_level = logical  # for physical + logical (CDC) replication
max_wal_senders = 10
wal_keep_size = 128MB
hot_standby = on
EOF

echo "host replication physical_rep 0.0.0.0/0 md5" >> "$PGDATA/pg_hba.conf"


echo "Primary configured."
```

Another Script will run on the replicaDB at the first time to start the streaming 

```sh
#!/usr/bin/env bash
set -e

PRIMARY_HOST=${PRIMARY_HOST:-"postgres_primary"}
PRIMARY_PORT=${PRIMARY_PORT:-5432}
REPL_USER=${REPL_USER:-"physical_rep"}
PGDATA=${PGDATA:-/var/lib/postgresql/data}


echo "Waiting for primary ${PRIMARY_HOST}:${PRIMARY_PORT}..."
until pg_isready -h "$PRIMARY_HOST" -p "$PRIMARY_PORT" -U postgres; do
  sleep 1
done
echo "Primary is ready."


if [ -s "$PGDATA/PG_VERSION" ]; then
  rm -rf "${PGDATA:?}/"*
fi


echo "Running pg_basebackup..."
export PGPASSWORD="${PGPASSWORD:-replicator_pass}"
pg_basebackup -h "$PRIMARY_HOST" -p $PRIMARY_PORT -D "$PGDATA" -U "$REPL_USER" -v -P -X stream -R

echo "Base backup complete. primary_conninfo and standby config written."


```

*There are a good resources about how that work at the end of the section.*

Now writes go to the Primary, and reads go to the replica. On top of that, we have Redis as a cache for posts. There’s also an invocation strategy for posts inside the code, and we expose endpoints so that when the user interacts with posts, the API gateway forwards those requests to the post service. The post service then executes the logic.

So the main components look like this:

- Post service logic (main business logic)
- Postgres master (write operations)
- Postgres replica (read operations)
- Redis cache (post caching & counters)

I’m aware that using both cache + read replica may be overkill in my situation, but this is just practice, not a real production system.


There’s also the counters part (likes count and comments count). The idea here is that I don’t want to update the DB on every user action because that would add unnecessary write load and we need a good luck scaling that under traffic spikes. So counters hit Redis to increments. Then every **N minutes**, I sync those counters back to the DB in batches. This means:

- Fast response for likes/comments actions (cache hit)
- Eventual consistency (it is not consistent at the moments but it will be) for counters which is totally acceptable.

The sync can be done via a scheduled job that reads the counters from Redis and writes them in bulk to Postgres.

## CDC + Kafka + Outbox Pattern


The Outbox Pattern is a design pattern used in distributed systems to ensure reliable communication between services when working with a database and a message broker (e.g., Kafka, RabbitMQ).

The core idea is:

- When a service performs a database update (e.g., creating a Post), it also writes an “outbox record” into a special table in the same database transaction.
- A separate worker reads the outbox table and publishes events/messages to the message broker.
- After successful publishing, the outbox record is marked as processed or removed.


Now there are multiple ways to implement outbox patterns. A good approach I found for this scenario is using CDC. So the flow is:

1. A post is written to Postgres.
2. CDC connector reads the change from the database WAL.
3. CDC connector publishes that record to a Kafka topic.
4. Other services can consume from that Kafka topic and react.

### How CDC Works

Postgres maintains a **Write-Ahead Log (WAL)**. Every write operation goes into the WAL before hitting disk, and this is actually what allows replication and durability. To make CDC work, we use a logical decoding slot that reads the WAL entries in a logical format.

So to simplify:

- WAL contains the insert/update/delete changes.
- Logical decoding extracts these changes into a readable form.
- CDC connector subscribes to these logical slots.
- The connector streams changes to Kafka topics.
- Consumers pick up the events.

To make that works some steps should be done:

- Create Publication on required table and user with required auth

```postgres
CREATE ROLE logical_rep WITH REPLICATION LOGIN PASSWORD 'YOUR_POSSWORD';

CREATE PUBLICATION my_pub FOR TABLE posts(user_id , post_id , created_at);

GRANT SELECT ON TABLE public.posts TO logical_rep;
```
- After Connecter is running. Create Json File with configs needed

```json
{
    "name": "debezium-config",
    "config": {
        "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
        "plugin.name": "pgoutput",
        "publication.name" :"my_pub",
        "slot.name": "logical_slot",
        "database.hostname": "postgres_primary",
        "database.port": "5432",
        "database.user": "logical_rep",
        "database.password": "YOUR_PASSWORD",
        "database.dbname" : "postdb",
        "topic.prefix": "post_service" ,
        "key.converter": "org.apache.kafka.connect.json.JsonConverter",
        "key.converter.schemas.enable": "false",
        "value.converter": "org.apache.kafka.connect.json.JsonConverter",
        "value.converter.schemas.enable": "false",
        "tombstones.on.delete": "false",
        "transforms": "unwrap",
        "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState"
    }
}
```

Then we can just send a curl request to the connecter with that data
```bash
curl -X POST -H "Content-Type: application/json" \
  --data @debezium.json \
  http://localhost:8083/connectors
```


These resources are really good about this part at all:
- [postgres-plugins](https://debezium.io/documentation/reference/3.4/postgres-plugins.html) - Debezium Documentations
- [postgres connecters](https://debezium.io/documentation/reference/3.4/connectors/postgresql.html) - Debezium Documentation
- [Postgres Replication](https://www.enterprisedb.com/postgres-tutorials/postgresql-replication-and-automatic-failover-tutorial) - postgresql-replication-and-automatic-failover-tutorial

- [Logical Replication in postgres ](https://www.postgresql.fastware.com/blog/inside-logical-replication-in-postgresql) - logical-replication-in-postgresql

- [OutBox Pattern](https://www.decodable.co/blog/revisiting-the-outbox-pattern#implementation-considerations) - outbox-pattern#implementation-considerations

- [Using CDC for Outbox](https://www.decodable.co/blog/the-wonders-of-postgres-logical-decoding-messages-for-cdc) - the-wonders-of-postgres-logical-decoding-messages-for-cdc

## Feed Service

Now it’s time to talk about the feed service. The main idea here is that earlier, when we talked about the post service, new posts were published into a Kafka topic. The feed service consumes those messages and builds the feed for users.


There are basically two strategies to build feeds at scale:

1. **Fan-out on write**
2. **Fan-out on read** (or hybrid, depending on the system)

### 1. Fan-out on Write

This strategy means that when the feed service consumes a `(post_id, user_id , time)`, a worker looks up all the followers of that user. For every follower, we push that (`post_id , time)` into their feed cache. We maintain a lightweight feed cache per user that stores only post IDs and time, so storage is not a big issue even for millions of users.


**Fanout on Write flow**:

- User X publishes a post.
- Feed service consumes `(post_id=X123, user_id=X , createdAt=25411)`.
- Feed service fetches the follower list for user X.
- Feed service splits the follower list into batches.
- Workers process these batches in parallel and push `(X123 , 25411)` into each follower’s feed cache.


This strategy works well for “normal users” because the fan-out volume is manageable and parallelization keeps things responsive.

There is an important practical detail here. As writing to the cache is **I/O-bound**, distributing the work across multiple workers improves performance. When fetching followers for an ordinary user, the follower count might be small (e.g., 50–200), but for high-followed accounts it can reach 10,000 or even 100,000. To avoid overloading a single worker, we dynamically spawn multiple workers. Each worker is assigned a fixed batch size (for example, 100 followers).

```go
// Go Code for processing every kafka message
	i := int32(0)
	var wg sync.WaitGroup
	for int(atomic.LoadInt32(&i)) < len(followers) {
		wg.Add(1)
		endIdx := min(int(atomic.LoadInt32(&i))+fw.workerThreshold, len(followers))
		go func(ids []string) {
			defer wg.Done()
			for _, id := range ids {
				fw.cache.Set(models.FeedItem{
					UserId:     id,
					PostId:     item.PostId,
					Created_at: item.Created_at,
				})
			}
		}(followers[i:endIdx])
		atomic.AddInt32(&i, int32(fw.workerThreshold))
	}
	wg.Wait()
```



I already test that, dividing the workload among multiple workers resulted in a **49× speedup** compared to a single worker. Additionally, allowing Go to utilize all CPU cores (instead of running all goroutines on a single thread) yielded a **29× performance improvement**.

```go
ali-mohamed@Ali-PC:~/projects/DMB/feed_service$ go test -bench=. 
goos: linux
goarch: amd64
pkg: github.com/alimx07/Distributed_Microservices_Backend/feed_service
cpu: AMD Ryzen 5 5600H with Radeon Graphics         
Benchmark_SingleWorker_5000Followers                           1        5312748326 ns/op
Benchmark_100PerWorker_5000Followers                          10         107598250 ns/op
Benchmark_100PerWorker_5000Followers_Parallel-12             338           3588947 ns/op
PASS
ok      github.com/alimx07/Distributed_Microservices_Backend/feed_service       13.840s
```

> **Note:** These tests use mocked APIs and simulated I/O. Real-world performance may differ due to things like network latency, system load, and production infra.


### 2. Fan-out on Read

The problem appears when the author has an extremely large follower count (More than a threshold we determine). Doing fan-out on write becomes too expensive because pushing IDs into millions of caches immediately at publish time can be slow and resource-heavy.

So for celebrity users, we switch to **fan-out on read**. The logic here is the opposite of fan-out on write:

- We do not push their post IDs into every follower’s feed cache at publish time.
- Instead, when a follower requests their feed, we check which celebrities they follow.
- For those celebrity authors, we pull their recent posts from celeb caches on demand.

#### Infinite Scroll and Cursoring

To support infinite scroll, a cursor is maintained. The cursor represents where the feed left off. So when the user scrolls:

- We use the cursor to fetch the next page of post IDs (and celebrity posts if needed).
- We merge and sort them.
- We return the next page with an updated cursor.

If the user refreshes the feed, the cursor can be reset to `null`, and we start fresh.

All of that is done using  **Redis** as its caching layer, relying on a **Sorted Set** structure with `CreatedAt` as the score. This allows efficient insertion and retrieval operations with a time complexity of **O(log n)**.



### Final summary of the Flow

- The post service publishes `(post_id, user_id , createdAt)` into Kafka.
- The feed service consumes those IDs.
- For normal authors:
  - Use **fan-out on write**
  - Fetch followers
  - Split into batches (e.g., 100 followers per batch)
  - Workers push IDs into follower feed caches
- For celebrities:
  - Use **fan-out on read**
  - Storing Celeb postIDs in celebCache to read on need
- During feed requests:
  - Merge cached feed + celebrity posts from caches
  - Sort by time
  - Get post objects from PostService
  - Get user metadata from UserServic
  - Return merged post objects + user metadata (up to pageSize) and updated cursor

> NOTE: The Feed Service shown here applies a basic post-sorting approach. Real production systems use more sophisticated ranking logic, but the goal here is simply to illustrate the concept.

[Algorithm powering the For You feed on X/Twitter](https://github.com/xai-org/x-algorithm) -  X Algorithm for feed creation

## User Service

The **User Service** manages user data and authentication flows. It handles both **User Management** and **Auth** responsibilities within one service and issues **JWT** tokens used across the system. It uses the same **Storage pattern** as Post Service.

### Responsibilities

- Store user information (email, password, profile data, etc.)
- Register new users
- Log users in and out
- Issue JWT tokens
- Validate authentication requests
- Cache recent user data


## Follow Service


The **Follow Service** manages the follower relationships between users. It models simple social graph behavior such as following and unfollowing. The service exposes APIs for other services to query or update follower data.

### Responsibilities

- Store follow relationships between users
- Allow users to follow and unfollow other users
- Query follower and following lists
- Check if User is Celebrity or not 


#### Future Database Considerations

- **Graph DB:** If complex graph queries or recommendations(in Feed Service) are introduced





