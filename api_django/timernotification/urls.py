from django.conf.urls import patterns, url
from timernotification import views

urlpatterns = patterns('',
    url(r'^fetion$', views.fetion),
    url(r'^email$', views.email),
)
