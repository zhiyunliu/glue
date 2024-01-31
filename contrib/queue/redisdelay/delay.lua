
local key = KEYS[1]
local max = ARGV[1]
local result = redis.call('zrangebyscore',key,0,max,'LIMIT',0,1)
if next(result) ~= nil and #result > 0 then
    local re = redis.call('zrem',key,unpack(result));
    if re > 0 then
        return result;
    end
else
    return {}
end
