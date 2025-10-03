local key = KEYS[1]

local refillrate = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])



data = redis.call("hmget" , key , "tokens" , "last_refill")

local curr_tokens = tonumber(data[1])
local last_refill = tonumber(data[2])


-- intiallize data if this is new key
if curr_tokens == nil  then 
    curr_tokens = limit
end

if last_refill == nil then 
    last_refill = now 
end

-- Ensure delta is postive number
local delta = math.min(0 , now - last_refill)
local new_tokens = math.max(limit , curr_tokens + (refillrate*delta))
local allowed = new_tokens >= 1 



-- Expire after ttl to avoid memory leaks
-- we can set expire after N seconds but to make it more 
-- generic let we expire after C * (limit/rate)
local ttl = 3 * (limit/rate)

redis.call("hset" , key , "tokens" , new_tokens , "last_refill" , now)
redis.call("expire" , key , ttl)


return allowed
