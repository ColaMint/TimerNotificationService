# -*- coding:utf-8 -*-
from django.db import models
from rest_framework import serializers
# Create your models here.

class FetionRequest(models.Model):
    ip=models.CharField(max_length=15)
    fetion_user=models.CharField(max_length=32)
    fetion_password=models.CharField(max_length=32)
    fetion_message=models.CharField(max_length=500)
    notification_time=models.CharField(max_length=20)
    create_time=models.DateTimeField(auto_now_add=True)
    is_done=models.BooleanField(default=False)
    class Meta:
        db_table="fetion_request"

class FetionRequestSerializer(serializers.ModelSerializer):
	class Meta:
		model=FetionRequest

class EmailRequest(models.Model):
    ip=models.CharField(max_length=15)
    email_from=models.CharField(max_length=32)
    email_subject=models.CharField(max_length=500)
    email_body=models.CharField(max_length=50000)
    email_type=models.PositiveSmallIntegerField()   # 1 : text/plain 2 : text/html
    email_to_users=models.CharField(max_length=1000)
    notification_time=models.CharField(max_length=20)
    create_time=models.DateTimeField(auto_now_add=True)
    is_done=models.BooleanField(default=False)
    class Meta:
        db_table="email_request"

class EmailRequestSerializer(serializers.ModelSerializer):
	class Meta:
		model = EmailRequest
