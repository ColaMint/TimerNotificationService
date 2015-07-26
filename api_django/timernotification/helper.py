from django.http import HttpRequest
from django.conf import settings 

import redis
import time
import json

from timernotification.models import *

def getRequestIP(request) :
    if request.META.has_key('HTTP_X_FORWARDED_FOR'):
        return request.META['HTTP_X_FORWARDED_FOR']
    else:
        return request.META['REMOTE_ADDR']

def getAntiKey(service, user) :
    return '%s-%s' % (service, user)

def nowTimestamp() :
    return int(time.time())

def getRedis(redisConfig) :
    pool = redis.ConnectionPool(host=redisConfig['HOST'],
                                port=redisConfig['PORT'],
                                db=redisConfig['DB'])
    return redis.Redis(connection_pool=pool)

def ensureRedisKeyType(r, key, keyType) :
    if r.type(key) != keyType :
        r.delete(key)


def getServiceCount(service, user, timelen) : 
    r = getRedis(settings.REDIS['ANTI'])
    key = getAntiKey(service, user)
    now = nowTimestamp()
    cnt = 0

    ensureRedisKeyType(r, key, 'list')

    for t in r.lrange(key, start = 0, end = -1) : 
        t = int(t)
        if t >= now - timelen and t <= now :
            cnt += 1
    return cnt

def updateServiceCount(service, user, timelen) :
    r = getRedis(settings.REDIS['ANTI'])
    key = getAntiKey(service, user)
    value = nowTimestamp()
    
    ensureRedisKeyType(r, key, 'list')

    r.lpush(key, value)
    r.expire(key, timelen)

def sendFetionTask(fetionRequest) :
    key = 'fetion-task-set'
    r = getRedis(settings.REDIS['TASK'])
    ensureRedisKeyType(r, key, 'zset')

    fetionRequestSerializer = FetionRequestSerializer(fetionRequest)
    r.zadd(key, json.dumps(fetionRequestSerializer.data), int(fetionRequest.notification_time))

def sendEmailTask(emailRequest) :
    key = 'email-task-set'
    r = getRedis(settings.REDIS['TASK'])
    ensureRedisKeyType(r, key, 'zset')

    emailRequestSerializer = EmailRequestSerializer(emailRequest)
    r.zadd(key, json.dumps(emailRequestSerializer.data), int(emailRequest.notification_time))
