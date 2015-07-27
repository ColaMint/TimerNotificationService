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
	"strconv"
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

func SendEmail(emailTask *EmailTask) bool {
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
		seelog.Errorf("[Send Email Failed] [EmailTask : %v] [Error : %v]", *emailTask, err)
		return false
	}

	seelog.Debugf("[Send Email Succeed] [EmailTask : %v]", *emailTask)
	return true
}

/* deprecated */
func GetAllNotDoneEmailTaskInDB() (emailTasks []*EmailTask) {
	emailTasks = make([]*EmailTask, 0)
	var db *sql.DB
	var stmt *sql.Stmt
	var err error
	LogError := func() {
		seelog.Errorf("[GetAllNotDoneEmailTaskInDB Failed] [ERROR : %v]", err)
	}

	if db, err = GetDBConnection(); err != nil {
		LogError()
		return
	}
	defer db.Close()

	if stmt, err = db.Prepare("SELECT `id`, `email_from`, `email_subject`, `email_body`, `email_type`, `email_to_users`, `notification_time` FROM `email_request` WHERE `is_done` = 0"); err != nil {
		LogError()
		return
	}
	defer stmt.Close()

	var rows *sql.Rows
	if rows, err = stmt.Query(); err != nil {
		LogError()
		return
	}
	defer rows.Close()

	for rows.Next() {
		var emailTask EmailTask
		rows.Scan(&emailTask.Id,
			&emailTask.From,
			&emailTask.Subject,
			&emailTask.Body,
			&emailTask.Type,
			&emailTask.ToUsers,
			&emailTask.NotificationTime)
		emailTasks = append(emailTasks, &emailTask)
		seelog.Debugf("[Load Email Task From DB] [EmailTask : %v]", emailTask)
	}

	return
}

func SetEmailTaskDone(emailTask *EmailTask) bool {
	var db *sql.DB
	var stmt *sql.Stmt
	var err error
	LogError := func() {
		seelog.Errorf("[SetEmailTaskDone Failed] [EmailTask : %v] [ERROR : %v]", *emailTask, err)
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

func FetchEmailTasksFromRedis() []interface{} {
	now := time.Now().Unix()
	emailTasks := make([]interface{}, 0)
	key := "email-task-set"
	conn := RedisPool.Get()
	if conn != nil {
		defer conn.Close()
		conn.Send("MULTI")
		conn.Send("ZRANGEBYSCORE", key, 0, now)
		conn.Send("ZREMRANGEBYSCORE", key, 0, now)
		queued, err := conn.Do("EXEC")
		if err == nil && queued != nil {
			jsonStrs, err := redis.Strings(queued.([]interface{})[0], nil)
			if err == nil {
				for _, jsonStr := range jsonStrs {
					seelog.Debugf("[Receive EmailTask From Redis] [Json : %v]", jsonStr)
					if emailTask, err := BuildEmailTaskFromJson(jsonStr); err == nil && emailTask != nil {
						if nt, err := strconv.Atoi(emailTask.NotificationTime); err == nil {
							/* 最多延迟一个小时发送 */
							delta := now - int64(nt)
							if delta < int64(time.Hour.Seconds()*1) {
								emailTasks = append(emailTasks, emailTask)
							} else {
								seelog.Debugf("[EmailTask Too Late] [Delta Seconds : %v][EmailTask : %v]", delta, *emailTask)
							}
						}
					}
				}
			}
		}
	}
	return emailTasks
}

type EmailTaskHandler struct {
}

func (*EmailTaskHandler) TaskName() string {
	return "Email"
}

func (*EmailTaskHandler) FetchTasks() []interface{} {
	return FetchEmailTasksFromRedis()
}

func (*EmailTaskHandler) HandleTask(task interface{}) bool {
	if emailTask, ok := task.(*EmailTask); ok {
		if SendEmail(emailTask) {
			SetEmailTaskDone(emailTask)
			return true
		} else {
			return false
		}
	} else {
		return false
	}

}

func (*EmailTaskHandler) TaskToString(task interface{}) string {
	if emailTask, ok := task.(*EmailTask); ok {
		return fmt.Sprintf("%v", *emailTask)
	} else {
		return "Unknown Task"
	}
}
