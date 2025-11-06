-- Check if a token exists in the denylist
-- KEYS[1]: token
-- Returns: 1 if token is in denylist (revoked), 0 if not found (valid)

local token_key = KEYS[1]

if redis.call("EXISTS", token_key) == 1 then
    return 1  -- Token is revoked
else
    return 0  -- Token is valid 
end
