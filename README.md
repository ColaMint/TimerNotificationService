# TimerNotificationService(未完成)

##简介
###TimerNotificationService项目通过API向服务器发送定时发送飞信/邮件的任务，并能限制用户调用频率。其中飞信限制为只能发送给自己。
---

##部署指南

#####Part-One：【api-django】 提供API给用户调用，并负责字段检查、频率限制等工作，最终把可信任的任务保存在redis中
*   安装pip
		
		sudo apt-get install python-pip
*   复制 /api_django/api_django/settings.py.example 一份到同目录下并命名为 settings.py
    
    修改其中数据库的配置和erdis的配置， 使其跟你的环境相匹配
    
    在你的数据库中创建一个名为TimerNotification的数据库，该数据库用于记录用户的请求

    另外，REDIS['ANTI']用于计算用户调用频率，REDIS['TASK']用于保存发送任务
        DATABASES = {
            'default': {
		    'ENGINE': 'django.db.backends.mysql',
		    'NAME': 'TimerNotification',
		    'USER': 'root',
		    'PASSWORD': '123456',
		    'HOST': '127.0.0.1',
		    'PORT': '3306',
		    }
	    }
        REDIS = {
	        'ANTI':{
		        'HOST': 'localhost',
		        'PORT': 6379,
		        'DB'  : 0,
	            },	
	    'TASK':{
		    'HOST': 'localhost',
		    'PORT': 6379,
		    'DB'  : 0,
	        }	
	    }
        
    还可以通过以下两个参数修改用户调用的频率
        MAX_FETION_COUNT_PER_HOUR = 10  
        MAX_EMAIL_COUNT_PER_HOUR = 10
        
*   进入/api_django，执行以下命令
        #安装项目依赖的python包
        pip install requirement.txt
        
        #自动建立数据库表
        python manage.py makemigrations
        python manage.py migrate

        #运行项目
        python manage.py runserver 0.0.0.0:80

*   添加定时任务API
        飞信任务
        URL     :   http://host:port/timer_notification/api/fetion
        METHOD  ：  POST
        BODY    ：  {   
                        'fetion_user'       :   '【飞信账号】'  ,                                                       
                        'fetion_password'   :   '【飞信密码】'  ,                                                                                   
                        'fetion_message'    :  '【飞信内容】'  ,                                                                                                                          
                        'notification_time' :   '【发送时间戳】 ,
                    }
        备注    :   飞信内容不能多于500个字符
                    发送时间不能比当前时间大30天

        邮件任务
        URL     :   http://host:port/timer_notification/api/email
        METHOD  ：  POST
        BODY    ：  {   'email_smpt'        :   '【smpt服务器】',                                                                                                      
                        'email_user'        :   '【邮箱账号】'  ,                                                                                                                                                                      
                        'email_password'    :  '【邮箱密码】'  ,                                                                                                                                                                         
                        'email_subject'     :   '【邮件标题】'  ,                                                                                                                                                                             
                        'email_body'        :   '【邮件内容】'  ,                                                                                                                                                                                           
                        'email_type'        :   '【内容类型】'  ,                                                                                                                                                                                       
                        'email_to_users',   :   '【收件人】'    ,                                                                                                                                                                                        
                        'notification_time' :   '【发送时间戳】',                                                                                                                                                                         
                      ) 
        备注    :   邮件标题不能多于500个字符
                    邮件内容不能多于50000个字符
                    内容类型可选的值有 :    1 --> text/plain    2 --> text/html
                    多个收件人用 ; 隔开， 总长度不能大于300字符
                    发送时间不能比当前时间大30天
