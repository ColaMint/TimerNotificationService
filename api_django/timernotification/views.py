from rest_framework.decorators import api_view,renderer_classes
from rest_framework.renderers import JSONRenderer
from rest_framework.response import Response

from django.shortcuts import render
from django.http import HttpRequest
from django.db import transaction

import json

from timernotification.helper import getRequestIP
from timernotification.models import FetionRequest, EmailRequest

# Create your views here.

@api_view(['POST'])
@renderer_classes((JSONRenderer,))
def fetion(request):
	recv_json = request.body

	try :
		recv_data = json.loads(recv_json)
	except Exception as e : 
		return errorResponse('invalid JSON data')

	ip = getRequestIP(request)
	if ip == None :
		return errorResponse('failed to parse your IP')

	requireFields = (	'fetion_user',
						'fetion_password',
						'fetion_message',
						'notification_time'	
					)

	for field in requireFields :
		try:
			checkField(recv_data, field)
		except FieldException as e:
			return errorResponse(e)
	
	fetionRequest = FetionRequest(	ip = ip,
									fetion_user = recv_data['fetion_user'],
									fetion_password = recv_data['fetion_password'],
									fetion_message = recv_data['fetion_message'],
									notification_time = recv_data['notification_time'],
								)
	try:
		with transaction.atomic():
			fetionRequest.save()
	except Exception as e:
		return errorResponse('database error')

	return successResponse()

@api_view(['POST'])
@renderer_classes((JSONRenderer,))
def email(request):
	recv_json = request.body

	try :
		recv_data = json.loads(recv_json)
	except Exception as e : 
		return errorResponse('invalid JSON data')

	ip = getRequestIP(request)
	if ip == None :
		return errorResponse('failed to parse your IP')

	requireFields = (	'email_smpt',
						'email_user',
						'email_password',
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
	
	emailRequest = EmailRequest(	ip = ip,
									email_smpt = recv_data['email_smpt'],
									email_user = recv_data['email_user'],
									email_password = recv_data['email_password'],
									email_subject = recv_data['email_subject'],
									email_body = recv_data['email_body'],
									email_type = recv_data['email_type'],
									email_to_users = recv_data['email_to_users'],
									notification_time = recv_data['notification_time'],
								)
	
	try:
		with transaction.atomic():
			emailRequest.save()
	except Exception as e:
		return errorResponse('database error')

	return successResponse()

def errorResponse(error, code = -1):
	message = '%s' % error
	return Response({"code" : code, "message" : message})

def successResponse(message = ''):
	return Response({"code" : 0, "message" : message})

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
