-- Add a token key to Redis
-- KEYS[1]: token 
-- ARGS[1] : ttl
local token_key = KEYS[1]
local ttl = tonumber(ARGV[1])
local value = "1"

return redis.call("SET", token_key , value , "EX" , ttl)
