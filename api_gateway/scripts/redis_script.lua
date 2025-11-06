local key = KEYS[1]

local refillrate = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])



local data = redis.call("hmget" , key , "tokens" , "last_refill")

local curr_tokens = tonumber(data[1])
local last_refill = tonumber(data[2])


-- intiallize data if this is new key
if curr_tokens == nil then 
    curr_tokens = limit
end

if last_refill == nil then 
    last_refill = now 
end

-- Ensure delta is postive number
local delta = math.max(0 , now - last_refill)
local new_tokens = math.min(limit , curr_tokens + (refillrate*delta))
local allowed = new_tokens >= 1 

if allowed then 
    new_tokens = new_tokens - 1
end

-- Expire after ttl to avoid memory leaks
-- we can set expire after N seconds but to make it more 
-- generic let we expire after C * (limit/rate)
local ttl = 3 * (limit/refillrate)

redis.call("hset" , key , "tokens" , new_tokens , "last_refill" , now)
redis.call("expire" , key , ttl)


return allowed
