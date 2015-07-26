# -*- coding:utf-8 -*-
from rest_framework.decorators import api_view,renderer_classes
from rest_framework.renderers import JSONRenderer
from rest_framework.response import Response

from django.shortcuts import render
from django.http import HttpRequest
from django.db import transaction
from django.conf import settings 
import json

from timernotification.helper import *
from timernotification.models import FetionRequest, EmailRequest

# Create your views here.

@api_view(['POST'])
@renderer_classes((JSONRenderer,))
def fetion(request):
    recv_json = request.body

    try :
        recv_data = json.loads(recv_json)
    except Exception as e : 
        return parseJsonErrprResponse()

    ip = getRequestIP(request)
    if ip == None :
        return parseIpErrorResponse() 

    requireFields = (
                        'fetion_user',
                        'fetion_password',
                        'fetion_message',
                        'notification_time'    
                    )

    for field in requireFields :
        try:
            checkField(recv_data, field)
        except FieldException as e:
            return errorResponse(e)
    
    if not checkNotificationTime(recv_data['notification_time']) :
        return invalidNotificationTime() 
    
    fetionRequest = FetionRequest(    ip = ip,
                                    fetion_user = recv_data['fetion_user'],
                                    fetion_password = recv_data['fetion_password'],
                                    fetion_message = recv_data['fetion_message'],
                                    notification_time = recv_data['notification_time'],
                                )

    # 限调用频率　每小时最多成功调用的次数 : settings.MAX_FETION_COUNT_PER_HOUR
    serviceName = 'fetion'
    maxTimesPerHour = settings.MAX_FETION_COUNT_PER_HOUR
    timelen = 3600
    if maxTimesPerHour >= 0 :
        docnt = getServiceCount(serviceName, fetionRequest.fetion_user, timelen)
        if docnt >= maxTimesPerHour :
            return reachServiceMaxTimesResponse()

    try:
        with transaction.atomic():
            fetionRequest.save()
    except Exception as e:
        return databaseErrorResponse()

    sendFetionTask(fetionRequest)
    updateServiceCount(serviceName, fetionRequest.fetion_user, timelen)

    return successResponse()

@api_view(['POST'])
@renderer_classes((JSONRenderer,))
def email(request):
    recv_json = request.body

    try :
        recv_data = json.loads(recv_json)
    except Exception as e : 
        return parseJsonErrprResponse() 

    ip = getRequestIP(request)
    if ip == None :
        return parseIpErrorResponse()

    requireFields = (    
                        'email_from',
                        'email_subject',
                        'email_body',
                        'email_type',
                        'email_to_users',
                        'notification_time'    
                    )
    for field in requireFields :
        try:
            checkField(recv_data, field)
        except FieldException as e:
            return errorResponse(e)
    
    if not checkNotificationTime(recv_data['notification_time']) :
        return invalidNotificationTime() 
    
    emailRequest = EmailRequest(    ip = ip,
                                    email_from = recv_data['email_from'],
                                    email_subject = recv_data['email_subject'],
                                    email_body = recv_data['email_body'],
                                    email_type = recv_data['email_type'],
                                    email_to_users = recv_data['email_to_users'],
                                    notification_time = recv_data['notification_time'],
                                )

    
    # 限调用频率　每小时最多成功调用的次数：settings.MAX_EMAIL_COUNT_PER_HOUR
    serviceName = 'email'
    maxTimesPerHour = settings.MAX_EMAIL_COUNT_PER_HOUR
    timelen = 3600
    if maxTimesPerHour >= 0 :
        docnt = getServiceCount(serviceName, emailRequest.email_from, timelen)
        if docnt >= maxTimesPerHour :
            return reachServiceMaxTimesResponse()

    try:
        with transaction.atomic():
            emailRequest.save()
    except Exception as e:
        return databaseErrorResponse()
    
    sendEmailTask(emailRequest)
    updateServiceCount(serviceName, emailRequest.email_from, timelen)

    return successResponse()

def errorResponse(error, code = -1):
    message = '%s' % error
    return Response({'code' : code, 'message' : message})

def successResponse(message = 'success'):
    return Response({'code' : 0, 'message' : message})

def parseJsonErrprResponse() :
    return errorResponse('invalid JSON data')

def parseIpErrorResponse() :
    return errorResponse('failed to parse your IP')
    
def databaseErrorResponse() :
    return errorResponse('database error')

def reachServiceMaxTimesResponse() :
    return errorResponse('reach max service times per hour')

def invalidNotificationTime() :
    return errorResponse('invalid notification time')

class FieldException(Exception):
    def __init__(self, mesg):
        self.mesg = mesg
    def __str__(self):
        return self.mesg

class MissingParamException(FieldException):
    def __init__(self, field):
        mesg = "missing param : %s" % field
        FieldException.__init__(self, mesg)
class ParamContentException():
    def __init__(self, field):
        mesg = "invalid value with param : %s" % field
        FieldException.__init__(self, mesg)
    
def checkField(data, field):
    if data.has_key(field) == False:
        raise MissingParamException(field)
    else:
        value = "%s" % field
        if len(value) == 0:
            raise ParamContentException(field)

def checkNotificationTime(notificationTime) :
    try : 
        notificationTime = int(notificationTime)
    except ValueError :
        return False
    
    now = nowTimestamp()
    return notificationTime >= now - 60 and notificationTime <= now + 3600 * 24 * 30
    
