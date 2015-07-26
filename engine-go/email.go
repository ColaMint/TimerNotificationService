package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gomail.v1"
	"strings"
	"time"
)

const (
	EMAIL_TYPE_PLAIN_TEXT = 1
	EMAIL_TYPE_HTML       = 2
)

type EmailTask struct {
	Id               int    `json:"id"`
	From             string `json:"email_from"`
	Subject          string `json:"email_subject"`
	Body             string `json:"email_body"`
	Type             int    `json:"email_type"`
	ToUsers          string `json:"email_to_users"`
	NotificationTime string `json:"notification_time"`
}

func SendEmail(emailTask EmailTask) bool {
	var mailer *gomail.Mailer
	var msg *gomail.Message
	emailConfig := &Config.EmailConfig

	msg = gomail.NewMessage()

	msg.SetHeader("From", emailTask.From)

	toUsers := strings.Split(emailTask.ToUsers, ",")
	msg.SetHeader("To", toUsers...)

	msg.SetHeader("Subject", emailTask.Subject)

	var emailType string
	switch emailTask.Type {
	case EMAIL_TYPE_PLAIN_TEXT:
		emailType = "text/plain"
	case EMAIL_TYPE_HTML:
		emailType = "text/html"
	}
	msg.SetBody(emailType, emailTask.Body)

	if emailConfig.SmtpTls {
		mailer = gomail.NewMailer(emailConfig.SmtpHost, emailConfig.SmtpUser, emailConfig.SmtpPassword, emailConfig.SmtpPort, gomail.SetTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	} else {
		mailer = gomail.NewMailer(emailConfig.SmtpHost, emailConfig.SmtpUser, emailConfig.SmtpPassword, emailConfig.SmtpPort)
	}

	if err := mailer.Send(msg); err != nil {
		logMsg := fmt.Sprintf("[Send Email Failed] [EmailTask : %v] [Error : %v]", emailTask, err)
		seelog.Error(logMsg)
		return false
	}

	return true
}

func SetEmailTaskDone(emailTask EmailTask) bool {
	var db *sql.DB
	var stmt *sql.Stmt
	var err error
	LogError := func() {
		seelog.Errorf("[Set EmailTaskDone Done Failed] [EmailTask : %v] [ERROR : %v]", emailTask, err)
	}

	if db, err = GetDBConnection(); err != nil {
		LogError()
		return false
	}
	defer db.Close()

	if stmt, err = db.Prepare("UPDATE `email_request` SET `is_done` = 1 WHERE `id` = ?"); err != nil {
		LogError()
		return false
	}
	defer stmt.Close()

	if _, err = stmt.Exec(emailTask.Id); err != nil {
		LogError()
		return false
	}

	return true
}

func BuildEmailTaskFromJson(jsonStr string) (*EmailTask, error) {
	var emailTask EmailTask
	var err error

	err = json.Unmarshal([]byte(jsonStr), &emailTask)
	return &emailTask, err
}

func FetchEmailTasksFromRedis() []*EmailTask {
	t := time.Now().Unix()
	emailTasks := make([]*EmailTask, 0)
	key := "email-task-set"
	conn := RedisPool.Get()
	if conn != nil {
		defer conn.Close()
		conn.Send("MULTI")
		conn.Send("ZRANGEBYSCORE", key, 0, t)
		conn.Send("ZREMRANGEBYSCORE", key, 0, t)
		queued, err := conn.Do("EXEC")
		if err == nil && queued != nil {
			jsonStrs, e := redis.Strings(queued.([]interface{})[0], err)
			if e == nil {
				for _, jsonStr := range jsonStrs {
					seelog.Infof("[Email Task Json] %v", jsonStr)
					if emailTask, e := BuildEmailTaskFromJson(jsonStr); e == nil && emailTask != nil {
						emailTasks = append(emailTasks, emailTask)
					}
				}
			}
		}
	}
	return emailTasks
}
